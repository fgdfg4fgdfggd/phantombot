package inits

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/zekroTJA/shinpuru/internal/commands"
	"github.com/zekroTJA/shinpuru/internal/core"
	"github.com/zekroTJA/shinpuru/internal/listeners"
	"github.com/zekroTJA/shinpuru/internal/util"
)

// InitConfig reads data from the config file and returns the
// resulting config object
// Errors will be logged to stderr and will cause an application exit.
func InitConfig(configLocation string, cfgParser core.ConfigParser) *core.Config {
	// Opening config file
	cfgFile, err := os.Open(configLocation)
	// If file not exist, try to create the preset config
	// file at the passed location. Then, the program will exit
	// and promt the user to edit the file
	if os.IsNotExist(err) {
		// Creating new config file
		cfgFile, err = os.Create(configLocation)
		if err != nil {
			util.Log.Fatal("Config file was not found and failed creating default config:", err)
		}
		// Using pased config pasrer to encode config
		err = cfgParser.Encode(cfgFile, core.NewDefaultConfig())
		if err != nil {
			util.Log.Fatal("Config file was not found and failed writing to new config file:", err)
		}
		util.Log.Fatal("Config file was not found. Created default config file. Please open it and enter your configuration.")
	} else if err != nil {
		util.Log.Fatal("Failed opening config file:", err)
	}

	// Use passed config parser to decode the data read from
	// the config file
	config, err := cfgParser.Decode(cfgFile)
	if err != nil {
		util.Log.Fatal("Failed decoding config file:", err)
	}

	// Checking the config version. If the version si outdated,
	// the user will be promted to update the config or re-create
	// it. After, the program exits.
	if config.Version < util.ConfigVersion {
		util.Log.Fatalf("Config file structure is outdated and must be re-created. Just rename your config and start the bot to recreate the latest valid version of the config.")
	}

	// If no owner ID was set, a warning will be print to stdout.
	if config.Discord.OwnerID == "" {
		util.Log.Warning("ATTENTION: Bot onwer ID is not set in config!",
			"You will not be identified as the owner of this bot so you will not have access to the owner-only commands!")
	}

	util.Log.Info("Config file loaded")

	return config
}

// InitDatabase initializes the connection to the
// set database depending on the config settings.
// Errors will be logged to stderr and will cause an application exit.
func InitDatabase(databaseCfg *core.ConfigDatabaseType) core.Database {
	var database core.Database
	var err error

	// Check which database type was used
	// in the config
	switch strings.ToLower(databaseCfg.Type) {
	// Connect to MySql/MariaDB
	case "mysql", "mariadb":
		database = new(core.MySql)
		err = database.Connect(databaseCfg.MySql)
	// Open SQlite3 File "connection"
	case "sqlite", "sqlite3":
		database = new(core.Sqlite)
		err = database.Connect(databaseCfg.Sqlite)
	}

	if err != nil {
		util.Log.Fatal("Failed connecting to database:", err)
	}

	util.Log.Info("Connected to database")

	return database
}

// InitCommandHandler initializes the command handler and command structs.
// Errors will be logged to stderr and will cause an application exit.
func InitCommandHandler(s *discordgo.Session, config *core.Config, database core.Database, twitchNotifyWorker *core.TwitchNotifyWorker) *commands.CmdHandler {
	cmdHandler := commands.NewCmdHandler(s, database, config, twitchNotifyWorker)

	// Init and register all commands
	cmdHandler.RegisterCommand(&commands.CmdHelp{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdPrefix{PermLvl: 10})
	cmdHandler.RegisterCommand(&commands.CmdPerms{PermLvl: 10})
	cmdHandler.RegisterCommand(&commands.CmdClear{PermLvl: 8})
	cmdHandler.RegisterCommand(&commands.CmdMvall{PermLvl: 5})
	cmdHandler.RegisterCommand(&commands.CmdInfo{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdSay{PermLvl: 3})
	cmdHandler.RegisterCommand(&commands.CmdQuote{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdGame{PermLvl: 999})
	cmdHandler.RegisterCommand(&commands.CmdAutorole{PermLvl: 9})
	cmdHandler.RegisterCommand(&commands.CmdReport{PermLvl: 5})
	cmdHandler.RegisterCommand(&commands.CmdModlog{PermLvl: 6})
	cmdHandler.RegisterCommand(&commands.CmdKick{PermLvl: 6})
	cmdHandler.RegisterCommand(&commands.CmdBan{PermLvl: 8})
	cmdHandler.RegisterCommand(&commands.CmdVote{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdProfile{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdId{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdMute{PermLvl: 4})
	cmdHandler.RegisterCommand(&commands.CmdMention{PermLvl: 4})
	cmdHandler.RegisterCommand(&commands.CmdNotify{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdVoicelog{PermLvl: 6})
	cmdHandler.RegisterCommand(&commands.CmdBug{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdStats{PermLvl: 0})
	cmdHandler.RegisterCommand(&commands.CmdTwitchNotify{PermLvl: 5})
	cmdHandler.RegisterCommand(&commands.CmdGhostping{PermLvl: 3})
	cmdHandler.RegisterCommand(&commands.CmdExec{PermLvl: 5})
	cmdHandler.RegisterCommand(&commands.CmdBackup{PermLvl: 9})

	// If release build ldFlag was not set to "TRUE",
	// also register the `test` command.
	if util.Release != "TRUE" {
		cmdHandler.RegisterCommand(&commands.CmdTest{})
	}

	// If custom permissions are set in the config,
	// set them for the registered commands here.
	// Also, the custom Bbt and guild owner permission
	// will be set here in the command handler.
	if config.Permissions != nil {
		cmdHandler.UpdateCommandPermissions(config.Permissions.CustomCmdPermissions)
		if config.Permissions.BotOwnerLevel > 0 {
			util.PermLvlBotOwner = config.Permissions.BotOwnerLevel
		}
		if config.Permissions.GuildOwnerLevel > 0 {
			util.PermLvlGuildOwner = config.Permissions.GuildOwnerLevel
		}
	}

	util.Log.Infof("%d commands registered", cmdHandler.GetCommandListLen())

	return cmdHandler
}

// InitDiscordBotSession adds all registered event handlers, initializes the discord bot session
// and runs the event loop, which will block the current thread.
// Errors will be logged to stderr and will cause an application exit.
func InitDiscordBotSession(session *discordgo.Session, config *core.Config, database core.Database, cmdHandler *commands.CmdHandler) {
	// Setting snowflake epoche like set in consts
	snowflake.Epoch = util.DefEpoche
	// Setting up snowflake nodes
	err := util.SetupSnowflakeNodes()
	if err != nil {
		util.Log.Fatal("Failed setting up snowflake nodes: ", err)
	}

	// Setting bot token in session
	session.Token = "Bot " + config.Discord.Token

	// Initializing and registering all command handlers
	session.AddHandler(listeners.NewListenerReady(config, database).Handler)
	session.AddHandler(listeners.NewListenerCmd(config, database, cmdHandler).Handler)
	session.AddHandler(listeners.NewListenerGuildJoin(config).Handler)
	session.AddHandler(listeners.NewListenerMemberAdd(database).Handler)
	session.AddHandler(listeners.NewListenerVote(database).Handler)
	session.AddHandler(listeners.NewListenerChannelCreate(database).Handler)
	session.AddHandler(listeners.NewListenerVoiceUpdate(database).Handler)
	session.AddHandler(listeners.NewListenerGhostPing(database, cmdHandler).Handler)
	session.AddHandler(listeners.NewListenerJdoodle(database).Handler)

	// Actually open discord bot session here
	// with initializing the event loop in a
	// go routine.
	err = session.Open()
	if err != nil {
		util.Log.Fatal("Failed connecting Discord bot session:", err)
	}
	// Ensuring the discord bot session will be quit
	// with a websocket signal that the bot account
	// will not be shown online until timeout.
	defer func() {
		util.Log.Info("Shutting down...")
		session.Close()
	}()

	// To block the current thread, now, a channel will be created
	// which wil lwait for a system signal, triggered by CTRL-C,
	// for example, to stop the event loop.
	util.Log.Info("Started event loop. Stop with CTRL-C...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// InitTwitchNotifyer initializes the twitch notifyer, if twitch OAuth2 App credentials
// were provided in the config, sets up the listener and the worker and sets up all
// cahnnels to notify from database.
// Erros will be displayed in stderr and will not cause the application to exit.
func InitTwitchNotifyer(session *discordgo.Session, config *core.Config, db core.Database) *core.TwitchNotifyWorker {
	if config.Etc == nil || config.Etc.TwitchAppID == "" {
		return nil
	}

	// Initialize the twitch notify handlers
	listener := listeners.NewListenerTwitchNotify(session, config, db)
	// Initialize the twitch notify worker which polls the twitch API
	// for triggering the event handlers
	tnw := core.NewTwitchNotifyWorker(config.Etc.TwitchAppID,
		listener.HandlerWentOnline, listener.HandlerWentOffline)

	// Getting all registered twitch notifies from database and register
	// them to the worker to listen for status change
	notifies, err := db.GetAllTwitchNotifies("")
	if err == nil {
		for _, notify := range notifies {
			if u, err := tnw.GetUser(notify.TwitchUserID, core.TwitchNotifyIdentID); err == nil {
				tnw.AddUser(u)
			}
		}
	} else {
		util.Log.Error("failed getting Twitch notify entreis: ", err)
	}

	return tnw
}
