package nested

import (
	"github.com/imdario/mergo"
)

// WithSmartMerge configures mergo to append slices but not overwrite existing
// values.
func WithSmartMerge(config *mergo.Config) {
	config.AppendSlice = true
	config.Overwrite = false
}
