package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"

	containers "github.com/AidanHarveyNelson/discord_docker_manager/containers"
)

// Bot parameters
var (
	GuildID        = os.Getenv("GUILD_ID")
	BotToken       = os.Getenv("BOT_TOKEN")
	RemoveCommands = true
)

var s *discordgo.Session
var docker = containers.NewDocker()

func init() {
	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func addServerCommand(server_name string, command *discordgo.ApplicationCommand) {

	options := discordgo.ApplicationCommandOption{
		Name:        server_name,
		Description: "Group of commands to control server: " + server_name,
		Options:     serverActions,
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
	}

	command.Options = append(command.Options, &options)
}

var (
	serverActions = []*discordgo.ApplicationCommandOption{
		{
			Name:        "start",
			Description: "Start command for the server",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "stop",
			Description: "Stop command for the server",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "restart",
			Description: "Restart command for the server",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "status",
			Description: "Status command for the server",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
	}

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "game-server",
			Description: "Group of commands that revole around managing serveers",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"game-server": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			content := ""

			serverName := options[0].Name
			childAction := options[0].Options
			serverID := serverInfo[serverName]
			switch childAction[0].Name {
			case "start":
				docker.StartContainer(serverID)
				content = "Server: " + serverName + " has been succesfully started"
			case "stop":
				docker.StopContainer(serverID)
				content = "Server: " + serverName + " has been succesfully stopped"
			case "status":
				status := docker.StatusContainer(serverID)
				content = "Server: " + serverName + " is currently in status \"" + status + "\""
			case "restart":
				docker.RestartContainer(serverID)
				content = "Server: " + serverName + " has been succesfully restarted"
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
	}

	serverInfo map[string]string
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	contList, err := docker.SearchContainers(20)
	if err != nil {
		fmt.Printf("Unable to retrieve container information. Stopping bot")
		s.Close()
		os.Exit(1)
	}

	// Get Game Server Command and add Docker Containers
	serverInfo = make(map[string]string, len(contList))
	for i, v := range contList {
		fmt.Printf("Current loop is: %d\n", i)
		fmt.Printf("Container ID is: %v\n", v.ID)
		fmt.Printf("Container Name is: %v\n", v.Names)
		curName := v.Names[0][1:]
		// Should probably be a map
		serverInfo[curName] = v.ID
		addServerCommand(curName, commands[0])
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	// Clean up commands on Server if Required
	if RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
