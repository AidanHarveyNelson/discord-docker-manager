package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID           = flag.String("guid", "", "Guild to run the bot against")
	BotToken          = flag.String("token", "", "Bot access token")
	DockerFilter      = flag.String("filter", "", "Filter for selecting correct docker containers")
	RemoveCommands    = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	StopNoPlayer      = flag.Bool("auto-stop", false, "Automatically stops all containers with no players")
	StopNoPlayerHours = flag.Int("auto-stop-hours", 0, "How many hours the server has to have no players before stopping")
)

var s *discordgo.Session
var docker *Docker

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	docker = NewDocker(*DockerFilter)
}

// Helper function to add Command Application Option per server
// Will then also add all child actions to the server
func getServerChoices(filter string) []*discordgo.ApplicationCommandOptionChoice {

	choices := []*discordgo.ApplicationCommandOptionChoice{}
	servList, err := docker.SearchContainers(20, filter)
	if err != nil {
		log.Println("Unable to find any containers")
		return choices
	}

	for _, server := range servList {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  server.Names[0][1:],
			Value: server.ID,
		})
	}
	log.Printf("Constructed the following automcomplete options %v", choices)
	return choices
}

var (
	serverNameOption = []*discordgo.ApplicationCommandOption{
		{
			Name:         "server-name",
			Description:  "Please specify which server you would like to interact with",
			Type:         discordgo.ApplicationCommandOptionString,
			Required:     true,
			Autocomplete: true,
		},
	}
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "game-server",
			Description: "Group of commands that revole around managing serveers",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "start",
					Description: "Start command for the server",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options:     serverNameOption,
				},
				{
					Name:        "stop",
					Description: "Stop command for the server",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options:     serverNameOption,
				},
				{
					Name:        "restart",
					Description: "Restart command for the server",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options:     serverNameOption,
				},
				{
					Name:        "status",
					Description: "Status command for the server",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options:     serverNameOption,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"game-server": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				options := i.ApplicationCommandData().Options
				content := ""
				subCommand := options[0].Options

				switch options[0].Name {
				case "start":
					content = "Request to start server has been recieved"
				case "stop":
					content = "Request to stop server has been recieved"
				case "status":
					content = "Request to get server status has been recieved"
				case "restart":
					content = "Request to restart server has been recieved"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: content,
					},
				})
				switch options[0].Name {
				case "start":
					docker.StartContainer(subCommand[0].Value.(string))
					content = "Server has been succesfully started"
				case "stop":
					go docker.StopContainer(subCommand[0].Value.(string))
					content = "Server has been succesfully stopped"
				case "status":
					status := docker.StatusContainer(subCommand[0].Value.(string))
					content = "Server is currently in status \"" + status + "\""
				case "restart":
					go docker.RestartContainer(subCommand[0].Value.(string))
					content = "Server has been succesfully restarted"
				}
				_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				if err != nil {
					s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: "Something went wrong",
					})
					return
				}

			case discordgo.InteractionApplicationCommandAutocomplete:
				data := i.ApplicationCommandData()
				var choices []*discordgo.ApplicationCommandOptionChoice
				// Set different server choices based on command
				switch data.Options[0].Name {
				case "start":
					choices = getServerChoices("status=exited,status=paused")
				case "stop":
					choices = getServerChoices("status=running")
				case "status":
					choices = getServerChoices("")
				case "restart":
					choices = getServerChoices("status=running")
				}
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionApplicationCommandAutocompleteResult,
					Data: &discordgo.InteractionResponseData{
						Choices: choices,
					},
				})
				if err != nil {
					panic(err)
				}
			}
		},
	}
	// serverInfo map[string]string
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

	// Keep a log of registered commands so we can clean up after the bot shut downs
	// This ensures that commands will always stay up to date
	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// Defer closure of connections
	defer s.Close()
	defer docker.Close()

	// Listen for OS Interrupt to run clean up
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	// Clean up commands on Server if Required
	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
