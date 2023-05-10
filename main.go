package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:  "git-semver",
		Usage: "Manage semantic version tags",
		Commands: []*cli.Command{
			&cmdGet,
			&cmdMajor,
			&cmdMinor,
			&cmdPatch,
			&cmdPreRelease,
			&cmdBuild,
		},
	}
	app.Run(os.Args)
}

var cmdPatch = cli.Command{
	Name:    "patch",
	Aliases: []string{"next"},
	Usage:   "Generate a tag for the next patch version",
	Flags: []cli.Flag{
		dryRunFlag(),
		preReleaseFlag(false),
		buildFlag(false),
	},
	Before: readConfig,
	Action: nextVersion("patch"),
}

var cmdMinor = cli.Command{
	Name:  "minor",
	Usage: "Generate a tag for the next minor version",
	Flags: []cli.Flag{
		dryRunFlag(),
		preReleaseFlag(false),
		buildFlag(false),
	},
	Before: readConfig,
	Action: nextVersion("minor"),
}

var cmdMajor = cli.Command{
	Name:  "major",
	Usage: "Generate a tag for the next major version",
	Flags: []cli.Flag{
		dryRunFlag(),
		preReleaseFlag(false),
		buildFlag(false),
	},
	Before: readConfig,
	Action: nextVersion("major"),
}

var cmdPreRelease = cli.Command{
	Name:  "pre-release",
	Usage: "Generate a tag for the specified pre-release",
	Flags: []cli.Flag{
		dryRunFlag(),
		preReleaseFlag(true),
		buildFlag(false),
	},
	Before: readConfig,
	Action: nextVersion(""),
}

var cmdBuild = cli.Command{
	Name:  "build",
	Usage: "Generate a tag for the specified build",
	Flags: []cli.Flag{
		dryRunFlag(),
		buildFlag(true),
	},
	Before: readConfig,
	Action: nextVersion(""),
}

var cmdGet = cli.Command{
	Name:   "get",
	Usage:  "Gets the current version tag",
	Before: readConfig,
	Action: func(c *cli.Context) error {
		conf := c.Context.Value(ctxKeyConfig).(*config)
		v, err := getVersion(conf.versionPrefix)
		if err != nil {
			return cli.Exit(err.Error(), 2)
		}
		if v == nil {
			return cli.Exit("no valid semver tags found", 2)
		}
		fmt.Println(conf.versionPrefix + v.String())
		return nil
	},
}

func nextVersion(inc string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		conf := c.Context.Value(ctxKeyConfig).(*config)
		v, err := getVersion(conf.versionPrefix)
		if err != nil {
			return cli.Exit(err.Error(), 2)
		}
		var newVer semver.Version
		if v == nil {
			// If there is no semver tag, create version 0.1.0
			v = semver.New(0, 1, 0, "", "")
			newVer = *v
		} else {
			// Otherwise, apply the specified increment to the existing version
			switch inc {
			case "patch":
				newVer = v.IncPatch()
			case "minor":
				newVer = v.IncMinor()
			case "major":
				newVer = v.IncMajor()
			default:
				newVer = *v
			}
		}
		if c.IsSet("pre-release") {
			newVer, err = newVer.SetPrerelease(c.String("pre-release"))
			if err != nil {
				return cli.Exit(fmt.Sprintf("error setting pre-release %q: %v", c.String("pre-release"), err), 3)
			}
		}
		if c.IsSet("build") {
			newVer, err = newVer.SetMetadata(c.String("build"))
			if err != nil {
				return cli.Exit(fmt.Sprintf("error setting build %q: %v", c.String("build"), err), 3)
			}
		}
		tagName := conf.versionPrefix + newVer.String()
		if !c.Bool("dryrun") {
			if err = createTag(tagName, conf.sign); err != nil {
				return cli.Exit(err.Error(), 3)
			}
		}
		fmt.Println(tagName)
		return nil
	}
}

func dryRunFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "dryrun",
		Aliases: []string{"d"},
		Usage:   "Show version without creating a git tag",
	}
}

func preReleaseFlag(required bool) cli.Flag {
	return &cli.StringFlag{
		Name:     "pre-release",
		Aliases:  []string{"p"},
		Usage:    "Sets the pre-release version component",
		Required: required,
	}
}

func buildFlag(required bool) cli.Flag {
	return &cli.StringFlag{
		Name:     "build",
		Aliases:  []string{"b"},
		Usage:    "Sets the build version component",
		Required: required,
	}
}

type ctxKey int

const ctxKeyConfig ctxKey = 0

type config struct {
	versionPrefix string
	sign          bool
}

func readConfig(c *cli.Context) error {
	paths := []string{".git-semver"}
	if p, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		paths = append(paths, filepath.Join(p, "git-semver"))
	}
	if p, ok := os.LookupEnv("HOME"); ok {
		paths = append(paths, filepath.Join(p, ".config", "git-semver"), filepath.Join(p, ".git-semver", "config"))
	}
	conf := config{}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return cli.Exit(err.Error(), 1)
		}
		defer f.Close()
		conf, err = parseConfig(f)
		if err != nil {
			return cli.Exit(fmt.Sprintf("error parsing %s: %v", p, err), 1)
		}
		break
	}
	c.Context = context.WithValue(c.Context, ctxKeyConfig, &conf)
	return nil
}

func parseConfig(f io.Reader) (config, error) {
	var conf config
	s := bufio.NewScanner(f)
	for s.Scan() {
		t := strings.TrimSpace(s.Text())
		if len(t) == 0 {
			continue
		}
		if strings.HasPrefix(t, "#") {
			continue
		}
		k, v, ok := strings.Cut(t, "=")
		if !ok {
			return conf, fmt.Errorf("error parsing %s: invalid syntax", t)
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if len(v) >= 2 && strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") {
			unquotedV, err := strconv.Unquote(v)
			if err != nil {
				return conf, fmt.Errorf("error parsing %s: invalid quoted string", t)
			}
			v = unquotedV
		}
		switch strings.ToUpper(k) {
		case "VERSION_PREFIX":
			conf.versionPrefix = v
		case "GIT_SIGN":
			b, err := strconv.ParseBool(v)
			if err != nil {
				return conf, fmt.Errorf("error parsing %s: invalid boolean value", t)
			}
			conf.sign = b
		default:
			return conf, fmt.Errorf("error parsing %s: unrecognized variable", t)
		}
	}
	return conf, s.Err()
}

func createTag(tagName string, sign bool) error {
	signFlag := "-a"
	if sign {
		signFlag = "-s"
	}
	out, err := exec.Command("git", "tag", signFlag, "-m", "Version "+tagName, tagName).CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return err
	}
	return nil
}

func getVersion(versionPrefix string) (*semver.Version, error) {
	out, err := exec.Command("git", "tag").CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return nil, err
	}
	var latest *semver.Version
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		tagName := s.Text()
		if strings.HasPrefix(tagName, versionPrefix) {
			v, err := semver.NewVersion(tagName)
			if err != nil {
				log.Printf("error parsing tag %q: %v", tagName, err)
				continue
			}
			if latest == nil || v.GreaterThan(latest) {
				latest = v
			}
		}
	}
	return latest, s.Err()
}
