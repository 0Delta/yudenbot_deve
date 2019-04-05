module github.com/0Delta/yudenbot_devel

go 1.12

require (
	github.com/0Delta/colog2slack v1.1.0
	github.com/0Delta/yudenbot_devel/discord v0.0.0
	github.com/0Delta/yudenbot_devel/eventdata v0.0.0
	github.com/0Delta/yudenbot_devel/twitter v0.0.0
	github.com/comail/colog v0.0.0-20160416085026-fba8e7b1f46c
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/0Delta/yudenbot_devel/eventdata => ./eventdata

replace github.com/0Delta/yudenbot_devel/twitter => ./twitter

replace github.com/0Delta/yudenbot_devel/discord => ./discord
