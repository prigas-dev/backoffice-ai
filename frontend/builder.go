package frontend

import (
	_ "embed"
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/phuslu/log"
)

type IBuilder interface {
	BuildFrontend() error
	Close()
}

type BuilderConfig struct {
	// folder where there is source files and a package.json
	Entrypoint string

	// where the build results will be outputed to
	DestinationFolder string
}

func NewBuilder(config *BuilderConfig) (IBuilder, error) {

	ctx, err := api.Context(api.BuildOptions{
		EntryPoints: []string{config.Entrypoint},
		Bundle:      true,
		Splitting:   true,
		Format:      api.FormatESModule,
		Outdir:      config.DestinationFolder,
		JSX:         api.JSXAutomatic,
		JSXDev:      true,
		Loader: map[string]api.Loader{
			".ts":   api.LoaderTS,
			".tsx":  api.LoaderTSX,
			".js":   api.LoaderJS,
			".jsx":  api.LoaderJSX,
			".css":  api.LoaderCSS,
			".json": api.LoaderJSON,
		},
		Write: true,
	})

	if err != nil {
		return nil, err
	}

	builder := &Builder{
		ctx: ctx,
	}

	return builder, nil
}

type Builder struct {
	ctx api.BuildContext
}

func (b *Builder) BuildFrontend() error {

	buildResult := b.ctx.Rebuild()

	if len(buildResult.Warnings) > 0 {
		log.Warn().Msgf("build warnings: %+v", buildResult.Warnings)
	}

	if len(buildResult.Errors) > 0 {
		log.Error().Msgf("build warnings: %+v", buildResult.Errors)
		return fmt.Errorf("build failed: %+v", buildResult.Errors)
	}

	return nil
}

func (b *Builder) Close() {
	b.ctx.Dispose()
}
