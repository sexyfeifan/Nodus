package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// 1) Add role field to fh_users (admin|user) and make every existing
		// user an admin so current installs keep working.
		users, err := app.FindCollectionByNameOrId("fh_users")
		if err != nil {
			return err
		}
		if users.Fields.GetByName("role") == nil {
			users.Fields.Add(&core.TextField{Name: "role", Required: false, Max: 20})
		}

		// Only superusers can create users; user creation goes through InitializeSystem.
		users.CreateRule = nil
		// Self-service update stays; deletes require admin.
		users.UpdateRule = types.Pointer("id = @request.auth.id")
		users.DeleteRule = types.Pointer(`@request.auth.role = "admin"`)
		if err := app.Save(users); err != nil {
			return err
		}

		records, err := app.FindAllRecords("fh_users")
		if err != nil {
			return err
		}
		for _, rec := range records {
			if rec.GetString("role") == "" {
				rec.Set("role", "admin")
				if err := app.Save(rec); err != nil {
					return err
				}
			}
		}

		// 2) Lock privileged collections behind the admin role.
		// fh_settings: deleting/updating it used to reopen /api/system/initialize
		// and let any authenticated user mint a new admin.
		settings, err := app.FindCollectionByNameOrId("fh_settings")
		if err != nil {
			return err
		}
		settings.CreateRule = nil
		settings.DeleteRule = nil
		settings.UpdateRule = types.Pointer(`@request.auth.role = "admin"`)
		if err := app.Save(settings); err != nil {
			return err
		}

		servers, err := app.FindCollectionByNameOrId("fh_servers")
		if err != nil {
			return err
		}
		servers.CreateRule = types.Pointer(`@request.auth.role = "admin"`)
		servers.UpdateRule = types.Pointer(`@request.auth.role = "admin"`)
		servers.DeleteRule = types.Pointer(`@request.auth.role = "admin"`)
		if err := app.Save(servers); err != nil {
			return err
		}

		proxies, err := app.FindCollectionByNameOrId("fh_proxies")
		if err != nil {
			return err
		}
		proxies.CreateRule = types.Pointer(`@request.auth.role = "admin"`)
		proxies.UpdateRule = types.Pointer(`@request.auth.role = "admin"`)
		proxies.DeleteRule = types.Pointer(`@request.auth.role = "admin"`)
		if err := app.Save(proxies); err != nil {
			return err
		}

		// 3) Speed up the latency "latest value per target" queries.
		if _, err := app.DB().NewQuery(
			"CREATE INDEX IF NOT EXISTS idx_metrics_raw_target_key_t ON fh_metrics_raw (targetId, metricKey, t)",
		).Execute(); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
