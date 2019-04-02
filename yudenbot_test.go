// yuden

package main

import (
	"context"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestYudenbot(t *testing.T) {
	type args struct {
		ctx      context.Context
		execList []Executor
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "bot test",
			args: args{
				ctx: nil,
				execList: []Executor{
					Executor{
						Name: "fizz",
						Fnc:  func(ctx context.Context) (err error) { log.Println("fizz"); return nil },
						Tick: 3 * time.Second,
					},
					Executor{
						Name: "buzz",
						Fnc:  func(ctx context.Context) (err error) { log.Println("buzz"); return nil },
						Tick: 5 * time.Second,
					},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			tt.args.ctx = ctx
			go YudenBot(tt.args.ctx, tt.args.execList)
			time.Sleep(25 * time.Second)
			cancel()
		})
	}
}

func Test_createAndPostDiscordChannel(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test 1",
			wantErr: true,
		},
		// TODO: Add test cases.
	}

	UpdateDiscordScedules([]discordschedule{})

	ctx := context.Background()
	buf, err := ioutil.ReadFile("./.config.yml")
	if err != nil {
		t.Fatal("Error while load config : ", err)
	}
	ctx = context.WithValue(ctx, config, buf)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createAndPostDiscordChannel(ctx); (err != nil) != tt.wantErr {
				t.Errorf("createAndPostDiscordChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
