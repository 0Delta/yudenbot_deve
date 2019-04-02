module github.com/0Delta/yudenbot_devel

go 1.12

require (
	github.com/0Delta/colog2slack v0.0.0-20190402144800-21ad6babcfa3
	github.com/0Delta/yudenbot_devel/discord v0.0.0
	github.com/0Delta/yudenbot_devel/eventdata v0.0.0
	github.com/0Delta/yudenbot_devel/twitter v0.0.0
	github.com/ChimeraCoder/anaconda v2.0.0+incompatible
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/0Delta/yudenbot_devel/eventdata => ./eventdata

replace github.com/0Delta/yudenbot_devel/twitter => ./twitter

replace github.com/0Delta/yudenbot_devel/discord => ./discord
