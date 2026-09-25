package main

import (
	"embed"
	"github.com/spf13/cobra"
	"io/fs"
	"log"

	"nodus/internal/application/dashboard"
	"nodus/internal/application/frpc"
	"nodus/internal/application/importer"
	"nodus/internal/application/monitoring"
	"nodus/internal/application/proxy"
	"nodus/internal/application/server"
	"nodus/internal/application/system"
	"nodus/internal/application/version"
	"nodus/internal/infrastructure/persistence"
	httphandler "nodus/internal/interfaces/http"
	_ "nodus/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

//go:embed all:pb_public
var pbPublicDir embed.FS

func main() {
	app := pocketbase.New()

	// enable auto creation of migration files when using the "migrate" command
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// Infrastructure (Repository implementations)
	serverRepo := persistence.NewServerRepository(app)
	proxyRepo := persistence.NewProxyRepository(app)

	// Application Services
	serverService := server.NewService(app, serverRepo)
	proxyService := proxy.NewService(app, proxyRepo)
	frpcService := frpc.NewService(app, serverRepo, proxyRepo)
	dashboardService := dashboard.NewService(app, serverService, proxyService)
	settingsService := system.NewService(app)
	geoService := monitoring.NewGeoService(app)
	metricsService := monitoring.NewMetricsService(app)
	metricsScheduler := monitoring.NewMetricsScheduler(app, metricsService)
	networkMonitorService := monitoring.NewMonitorService(app, geoService, metricsService)
	importService := importer.NewService(app)
	versionService := version.NewService(app)

	// HTTP Handlers
	frpcHandler := httphandler.NewFrpcHandler(app, frpcService)
	versionHandler := httphandler.NewVersionHandler(app, versionService)
	dashboardHandler := httphandler.NewDashboardHandler(dashboardService)
	systemHandler := httphandler.NewSystemHandler(app, settingsService)
	importHandler := httphandler.NewImportHandler(app, importService)
	serverHandler := httphandler.NewServerHandler(app, metricsService)

	// Hook: Prevent privilege escalation — self-service user updates must not
	// change the role field (collection rules cannot restrict individual fields).
	app.OnRecordUpdateRequest("fh_users").BindFunc(func(e *core.RecordRequestEvent) error {
		newRole := e.Record.GetString("role")
		oldRole := e.Record.Original().GetString("role")
		if newRole != oldRole {
			isAdmin := e.Auth != nil && e.Auth.GetString("role") == "admin"
			if !isAdmin {
				return apis.NewForbiddenError("Only admins can change user roles", nil)
			}
		}
		return e.Next()
	})

	// Hook: Enforce unique proxy names within the same server so frp config
	// stays valid (frp requires unique proxy names per client).
	checkProxyNameUnique := func(e *core.RecordRequestEvent) error {
		name := e.Record.GetString("name")
		serverId := e.Record.GetString("serverId")
		if name == "" || serverId == "" {
			return e.Next()
		}
		filter := "name = {:name} && serverId = {:serverId}"
		params := map[string]any{"name": name, "serverId": serverId}
		if e.Record.Id != "" {
			filter += " && id != {:id}"
			params["id"] = e.Record.Id
		}
		_, err := e.App.FindFirstRecordByFilter("fh_proxies", filter, params)
		if err == nil {
			return apis.NewBadRequestError("A proxy with this name already exists on the selected server", nil)
		}
		return e.Next()
	}
	app.OnRecordCreateRequest("fh_proxies").BindFunc(checkProxyNameUnique)
	app.OnRecordUpdateRequest("fh_proxies").BindFunc(checkProxyNameUnique)

	// Hook: Reload frpc when proxy is updated
	app.OnRecordAfterUpdateSuccess("fh_proxies").BindFunc(func(e *core.RecordEvent) error {
		serverId := e.Record.GetString("serverId")
		if serverId != "" {
			// Check if the server is currently running
			if frpcService.IsServerRunning(serverId) {
				app.Logger().Info("Proxy updated, reloading frpc configuration", "proxyId", e.Record.Id, "serverId", serverId)
				if err := frpcService.ReloadFrpc(&serverId); err != nil {
					app.Logger().Error("Failed to reload frpc after proxy update", "error", err, "serverId", serverId)
				}
			}
		}
		return e.Next()
	})

	// Hook: Reload frpc when a proxy is deleted so the change takes effect immediately
	app.OnRecordAfterDeleteSuccess("fh_proxies").BindFunc(func(e *core.RecordEvent) error {
		serverId := e.Record.GetString("serverId")
		if serverId != "" && frpcService.IsServerRunning(serverId) {
			app.Logger().Info("Proxy deleted, reloading frpc configuration", "proxyId", e.Record.Id, "serverId", serverId)
			if err := frpcService.ReloadFrpc(&serverId); err != nil {
				app.Logger().Error("Failed to reload frpc after proxy delete", "error", err, "serverId", serverId)
			}
		}
		return e.Next()
	})

	// Hook: Stop frpc when a server is deleted to avoid orphaned processes
	app.OnRecordAfterDeleteSuccess("fh_servers").BindFunc(func(e *core.RecordEvent) error {
		serverId := e.Record.Id
		if frpcService.IsServerRunning(serverId) {
			app.Logger().Info("Server deleted, terminating frpc", "serverId", serverId)
			if err := frpcService.TerminateFrpc(&serverId); err != nil {
				app.Logger().Error("Failed to terminate frpc after server delete", "error", err, "serverId", serverId)
			}
		}
		metricsService.InvalidateTarget(serverId)
		return e.Next()
	})

	// Hook: Gracefully close all frpc clients on app shutdown
	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		frpcService.CloseAll()
		return e.Next()
	})

	// register custom routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Reset all server statuses to stopped on startup
		serverService.ResetAllServerStatus()
		// Reset all proxy statuses to offline on startup
		proxyService.ResetAllProxyBootStatus()

		// Auto-start servers with autoConnection enabled (non-blocking)
		go frpcService.AutoStartServers()

		// Start network monitoring service using intervals from settings
		latencyInterval, geoInterval := settingsService.GetMonitoringIntervals()
		networkMonitorService.Start(latencyInterval, geoInterval)
		metricsScheduler.Register()

		// Register routes for each module
		frpcHandler.RegisterHandlers(e)
		versionHandler.RegisterHandlers(e)
		dashboardHandler.RegisterHandlers(e)
		systemHandler.RegisterHandlers(e)
		importHandler.RegisterHandlers(e)
		serverHandler.RegisterHandlers(e)

		// Serve static files from embedded FS
		content, err := fs.Sub(pbPublicDir, "pb_public")
		if err != nil {
			return err
		}

		// Check if index.html exists to ensure frontend is properly embedded
		if _, err := content.Open("index.html"); err != nil {
			log.Printf("WARNING: index.html not found in embedded assets! Frontend may not load. Error: %v", err)
		}

		e.Router.GET("/{path...}", apis.Static(content, true))

		return e.Next()
	})

	// Overwrite the default serve command to suppress the default startup banner
	// and use our custom logs
	for _, cmd := range app.RootCmd.Commands() {
		if cmd.Use == "serve" {
			cmd.Run = nil
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				domains := args

				httpAddr, _ := cmd.Flags().GetString("http")
				httpsAddr, _ := cmd.Flags().GetString("https")
				origins, _ := cmd.Flags().GetStringSlice("origins")

				// replicate default logic from pocketbase
				if httpAddr == "" {
					if len(domains) > 0 {
						httpAddr = "0.0.0.0:80"
					} else {
						httpAddr = "127.0.0.1:8090"
					}
				}

				if httpsAddr == "" {
					if len(domains) > 0 {
						httpsAddr = "0.0.0.0:443"
					}
				}

				log.Println("Starting PocketBase server...")
				log.Println("----------------------------------------------------------------")
				log.Println("Frontend: http://" + httpAddr + "/")
				log.Println("Admin UI: http://" + httpAddr + "/_/")
				log.Println("----------------------------------------------------------------")

				return apis.Serve(app, apis.ServeConfig{
					HttpAddr:           httpAddr,
					HttpsAddr:          httpsAddr,
					ShowStartBanner:    false,
					AllowedOrigins:     origins,
					CertificateDomains: domains,
				})
			}
			break
		}
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
