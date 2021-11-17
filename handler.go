package router

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type Router struct {
	commands []command
	handlers map[discord.CommandID]Handler
	appID    discord.AppID
	c        *api.Client
}

func NewRouter(client *api.Client) (*Router, error) {
	r := &Router{
		handlers: map[discord.CommandID]Handler{},
		c:        client,
	}
	app, err := client.CurrentApplication()
	if err != nil {
		return nil, err
	}
	r.appID = app.ID
	return r, nil
}

func (r *Router) HandleInteraction(interaction discord.InteractionEvent) error {
	if interaction.Data.InteractionType() != discord.CommandInteractionType {
		return nil
	}
	data := interaction.Data.(*discord.CommandInteraction)
	if handler, ok := r.handlers[data.ID]; ok {
		context := &Context{
			c:           r.c,
			appID:       r.appID,
			Interaction: interaction,
			Command:     *data,
		}
		return handler(context)
	} else {
		return nil
	}
}

func (r *Router) AddCommand(cmd discord.Command, handler Handler) {
	r.commands = append(r.commands, command{
		cmd:     cmd,
		handler: handler,
	})
}

func (r *Router) RegisterCommands() error {
	return r.registerCommands(0)
}

func (r *Router) RegisterGuildCommands(gid discord.GuildID) error {
	if !gid.IsValid() {
		return fmt.Errorf("invalid guild ID %d", gid)
	}
	return r.registerCommands(gid)
}

func (r *Router) registerCommands(gid discord.GuildID) error {
	cmds := make([]discord.Command, len(r.commands))
	for i, c := range r.commands {
		cmds[i] = c.cmd
	}
	var err error
	var registered []discord.Command
	if gid == 0 {
		registered, err = r.c.BulkOverwriteCommands(r.appID, cmds)
	} else {
		registered, err = r.c.BulkOverwriteGuildCommands(r.appID, gid, cmds)
	}
	if err != nil {
		return err
	}
	r.handlers = make(map[discord.CommandID]Handler, len(registered))
	for _, cmd := range r.commands {
		for _, dcmd := range registered {
			if cmd.cmd.Name == dcmd.Name {
				r.handlers[dcmd.ID] = cmd.handler
				break
			}
		}
	}
	return nil
}

type Context struct {
	Interaction discord.InteractionEvent
	Command     discord.CommandInteraction

	c     *api.Client
	appID discord.AppID
}

func (c *Context) Response() (*discord.Message, error) {
	return c.c.InteractionResponse(c.appID, c.Interaction.Token)
}

func (c *Context) EditResponse(content string, embeds ...discord.Embed) (*discord.Message, error) {
	return c.EditResponseComplex(api.EditInteractionResponseData{
		Content: option.NewNullableString(content),
		Embeds:  &embeds,
	})
}

func (c *Context) EditResponseComplex(data api.EditInteractionResponseData) (*discord.Message, error) {
	return c.c.EditInteractionResponse(c.appID, c.Interaction.Token, data)
}

func (c *Context) Respond(content string, embeds ...discord.Embed) error {
	return c.RespondComplex(api.InteractionResponseData{
		Content: option.NewNullableString(content),
		Embeds:  &embeds,
	})
}

func (c *Context) RespondComplex(data api.InteractionResponseData) error {
	return c.respondInteraction(api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &data,
	})
}

func (c *Context) Defer() error {
	return c.respondInteraction(api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
	})
}

func (c *Context) Followup(content string, embeds ...discord.Embed) (*discord.Message, error) {
	return c.FollowupComplex(api.InteractionResponseData{
		Content: option.NewNullableString(content),
		Embeds:  &embeds,
	})
}

func (c *Context) FollowupComplex(data api.InteractionResponseData) (*discord.Message, error) {
	return c.c.CreateInteractionFollowup(c.appID, c.Interaction.Token, data)
}

func (c *Context) respondInteraction(resp api.InteractionResponse) error {
	return c.c.RespondInteraction(c.Interaction.ID, c.Interaction.Token, resp)
}

type Handler func(ctx *Context) error

type command struct {
	cmd     discord.Command
	handler Handler
}
