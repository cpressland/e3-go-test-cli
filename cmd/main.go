package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cpressland/e3-go-test-api/contracts"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Get a user",
				Action: cliGetUser,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "id",
						Required: true,
					},
				},
			},
			{
				Name:   "add",
				Usage:  "Add a user",
				Action: cliAddUser,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "email",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "phone",
						Required: true,
					},
				},
			},
			{
				Name:   "list",
				Usage:  "List all users",
				Action: cliListUsers,
			},
			{
				Name:   "delete",
				Usage:  "Delete a user",
				Action: cliDeleteUser,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "id",
						Required: true,
					},
				},
			},
			{
				Name:   "update",
				Usage:  "Update a user",
				Action: cliUpdateUser,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "id",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "username",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "email",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "phone",
						Required: false,
					},
				},
			},
		},
	}
	_ = app.Run(os.Args)
}

func getUser(id int) (contracts.GetUserResponse, error) {
	req, err := http.Get(fmt.Sprintf("https://api.e3.cpressland.io/users/%d", id))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to make request")
		return contracts.GetUserResponse{}, err
	}
	defer req.Body.Close()
	var resp contracts.GetUserResponse
	err = json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode response")
		return contracts.GetUserResponse{}, err
	}
	return resp, nil
}

func cliGetUser(cliCtx *cli.Context) error {
	user, err := getUser(cliCtx.Int("id"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get user")
		return err
	}
	log.Info().Interface("user", user).Send()
	return nil
}

func cliUpdateUser(cliCtx *cli.Context) error {
	existing_record, _ := getUser(cliCtx.Int("id"))
	var username, email, phone string

	if cliCtx.String("username") == "" {
		username = existing_record.Username
	} else {
		username = cliCtx.String("username")
	}
	if cliCtx.String("email") == "" {
		email = existing_record.Email
	} else {
		email = cliCtx.String("email")
	}
	if cliCtx.String("phone") == "" {
		phone = existing_record.Telephone
	} else {
		phone = cliCtx.String("phone")
	}

	user := contracts.UpdateUserRequest{
		Username:  username,
		Email:     email,
		Telephone: phone,
	}
	payload, err := json.Marshal(user)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal user")
	}
	url := fmt.Sprintf("https://api.e3.cpressland.io/users/%d", cliCtx.Int("id"))
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create request")
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to make request")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		log.Fatal().Str("err", string(data)).Msg("Failed to update user")
		return fmt.Errorf("failed to update user: %s", string(data))
	}
	log.Info().Msg("User updated")
	return nil
}

func cliAddUser(cliCtx *cli.Context) error {
	user := contracts.CreateUserRequest{
		Username:  cliCtx.String("username"),
		Email:     cliCtx.String("email"),
		Telephone: cliCtx.String("phone"),
	}
	payload, err := json.Marshal(user)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal user")
		return err
	}
	req, err := http.Post("https://api.e3.cpressland.io/users", "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to make request")
		return err
	}
	defer req.Body.Close()
	if req.StatusCode > 299 {
		data, _ := io.ReadAll(req.Body)
		log.Fatal().Str("err", string(data)).Msg("Failed to create user")
		return fmt.Errorf("failed to create user: %s", string(data))
	}
	var resp contracts.CreateUserResponse
	err = json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode response")
		return err
	}
	log.Info().Msgf("%#v\n", resp)
	return nil
}

func cliListUsers(cliCtx *cli.Context) error {
	req, err := http.Get("https://api.e3.cpressland.io/users")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to make request")
		return err
	}
	defer req.Body.Close()
	if req.StatusCode > 299 {
		data, _ := io.ReadAll(req.Body)
		log.Fatal().Str("err", string(data)).Msg("Failed to list users")
		return fmt.Errorf("failed to list users: %s", string(data))
	}
	var resp []contracts.GetUserResponse
	err = json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode response")
		return err
	}
	for _, user := range resp {
		log.Info().Interface("user", user).Send()
	}
	return nil
}

func cliDeleteUser(cliCtx *cli.Context) error {
	url := fmt.Sprintf("https://api.e3.cpressland.io/users/%d", cliCtx.Int("id"))
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create request")
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to make request")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		log.Fatal().Str("err", string(data)).Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %s", string(data))
	}
	log.Info().Msg("User deleted")
	return nil
}
