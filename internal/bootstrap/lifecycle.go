package bootstrap

import "context"

func (a *App) Start(ctx context.Context) error {
	a.log.InfoContext(ctx, "starting api service", "address", a.config.App.Address(), "env", a.config.App.Env)
	a.dependencies.Database.LogStatus(ctx)
	return a.server.Start()
}

func (a *App) Stop(ctx context.Context) error {
	a.log.InfoContext(ctx, "stopping api service")
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	return a.dependencies.Database.Close()
}
