package main

import (
	"context"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"go.samhza.com/discord/router"
)

func main() {
	guildID := discord.GuildID(mustSnowflake("GUILD_ID"))
	s, err := state.New("Bot " + os.Getenv("BOT_TOKEN"))
	s.AddIntents(0)
	if err != nil {
		log.Fatalln(err)
	}
	router, err := router.NewRouter(s.Client)
	if err != nil {
		log.Fatalln(err)
	}
	router.AddCommand(discord.Command{
		Name:        "ping",
		Description: "sends pong :3",
	}, ping)
	s.AddHandler(func(evt *gateway.InteractionCreateEvent) {
		router.HandleInteraction(evt.InteractionEvent)
	})
	err = router.RegisterGuildCommands(guildID)
	if err != nil {
		log.Fatalln(err)
	}
	err = s.Open(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	select {}
}

func mustSnowflake(env string) discord.Snowflake {
	str := os.Getenv(env)
	n, err := discord.ParseSnowflake(str)
	if err != nil {
		log.Fatalf("error parsing snowflake '%s': %s\n", str, err)
	}
	if !n.IsValid() {
		log.Fatalf("invalid snowflake '%s': %s\n", str, err)
	}
	return n
}

func ping(ctx *router.Context) error {
	return ctx.Respond("pong")
}
