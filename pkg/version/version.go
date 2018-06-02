//go:generate ../../hack/codegen-version.sh -o version.gen.go
package version

import (
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
)

var (
	buildDate string
)

func init() {
	if len(buildDate) != 0 {
		_, err := time.Parse(dateFormat, buildDate)
		if err != nil {
			panic(err)
		}
	}
}

func VersionInfo() Info {
	gd, _ := ptypes.TimestampProto(GitDate)
	bd, _ := ptypes.TimestampProto(BuildDate())
	return Info{
		GitCommit: GitCommit,
		Version:   Version,
		GitDate:   gd,
		BuildDate: bd,
	}
}

func (m Info) Json() string {
	v, err := (&jsonpb.Marshaler{}).MarshalToString(&m)
	if err != nil {
		panic(err)
	}
	return string(v)
}

func BuildDate() time.Time {
	if len(buildDate) == 0 {
		return GitDate
	}
	t, _ := time.Parse(dateFormat, buildDate)
	return t
}
