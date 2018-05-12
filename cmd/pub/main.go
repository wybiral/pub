package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli"
	"github.com/wybiral/pub/internal/api/private"
	"github.com/wybiral/pub/internal/api/public"
	"github.com/wybiral/pub/internal/app"
	"github.com/wybiral/pub/internal/model"
	"log"
	"os"
	"strings"
)

const version = "0.0.1"

func main() {
	c := cli.NewApp()
	cli.HelpFlag = cli.StringFlag{Hidden: true}
	cli.VersionFlag = cli.StringFlag{Hidden: true}
	c.Version = version
	c.Usage = "p2p publishing platform"
	c.Commands = []cli.Command{
		// create command
		cli.Command{
			Name:      "create",
			ArgsUsage: "DATABASE",
			Usage:     "Create identity",
			Action:    createIdentity,
			Flags:     []cli.Flag{},
		},
		// start command
		cli.Command{
			Name:      "start",
			ArgsUsage: "DATABASE",
			Usage:     "Start server",
			Action:    startServer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "socks-host",
					Value: "127.0.0.1",
					Usage: "Tor SOCKS host",
				},
				cli.IntFlag{
					Name:  "socks-port",
					Value: 9050,
					Usage: "Tor SOCKS port",
				},
				cli.StringFlag{
					Name:  "control-host",
					Value: "127.0.0.1",
					Usage: "Tor controller host",
				},
				cli.IntFlag{
					Name:  "control-port",
					Value: 9051,
					Usage: "Tor controller port",
				},
				cli.StringFlag{
					Name:  "control-password",
					Value: "",
					Usage: "Tor controller password",
				},
			},
		},
		// help command
		cli.Command{
			Name:      "help",
			Usage:     "Shows all commands or help for one command",
			ArgsUsage: "[command]",
			Action: func(c *cli.Context) {
				args := c.Args()
				if args.Present() {
					cli.ShowCommandHelp(c, args.First())
					return
				}
				cli.ShowAppHelp(c)
			},
		},
		// version command
		cli.Command{
			Name:  "version",
			Usage: "Print version",
			Action: func(ctx *cli.Context) {
				fmt.Println(c.Version)
			},
		},
	}
	err := c.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createIdentity(c *cli.Context) {
	args := c.Args()
	if len(args) != 1 {
		// Show help if no DB path supplied
		cli.ShowCommandHelp(c, "create")
		return
	}
	dbPath := normalizeDBPath(args[0])
	reader := bufio.NewReader(os.Stdin)
	// Get name
	fmt.Print("Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSuffix(name, "\n")
	// Get bio
	fmt.Print("About: ")
	about, _ := reader.ReadString('\n')
	about = strings.TrimSuffix(about, "\n")
	// Get DB model
	model, err := model.NewModel(dbPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Create self identity
	_, err = model.CreateSelf(name, about)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func startServer(c *cli.Context) {
	args := c.Args()
	if len(args) != 1 {
		// Show help if no DB path supplied
		cli.ShowCommandHelp(c, "start")
		return
	}
	dbPath := normalizeDBPath(args[0])
	// Setup app config
	config := app.NewDefaultConfig()
	config.DatabasePath = dbPath
	config.TorConfig.SocksHost = c.String("socks-host")
	config.TorConfig.SocksPort = c.Int("socks-port")
	config.TorConfig.ControlHost = c.String("control-host")
	config.TorConfig.ControlPort = c.Int("control-port")
	config.TorConfig.ControlPassword = c.String("control-password")
	// Create app
	a, err := app.NewApp(config)
	if err != nil {
		log.Fatal(err)
	}
	// Start APIs
	go public.StartApi(a)
	private.StartApi(a)
}

// Normalize DB path to reduce human error on input.
func normalizeDBPath(dbPath string) string {
	dbPath = strings.ToLower(dbPath)
	if !strings.HasSuffix(dbPath, ".db") {
		dbPath = dbPath + ".db"
	}
	return dbPath
}
