package images

import "github.com/defenseunicorns/zarf/src/types"

type ImgConfig struct {
	TarballPath string

	ImgList []string

	RegInfo types.RegistryInfo

	NoChecksum bool
}
