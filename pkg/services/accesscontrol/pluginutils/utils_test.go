package pluginutils

import (
	"testing"

	"github.com/grafana/grafana/pkg/plugins"
	ac "github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/stretchr/testify/require"
)

func TestToRegistrations(t *testing.T) {
	tests := []struct {
		name string
		regs []plugins.RoleRegistration
		want []ac.RoleRegistration
	}{
		{
			name: "no registration",
			regs: nil,
			want: []ac.RoleRegistration{},
		},
		{
			name: "registration gets converted successfully",
			regs: []plugins.RoleRegistration{
				{
					Role: plugins.Role{
						Name:        "test:name",
						DisplayName: "Test",
						Description: "Test",
						Permissions: []plugins.Permission{
							{Action: "test:action"},
							{Action: "test:action", Scope: "test:scope"},
						},
					},
					Grants: []string{"Admin", "Editor"},
				},
				{
					Role: plugins.Role{
						Name:        "test:name",
						Permissions: []plugins.Permission{},
					},
				},
			},
			want: []ac.RoleRegistration{
				{
					Role: ac.RoleDTO{
						Version:     1,
						Name:        "test:name",
						DisplayName: "Test",
						Description: "Test",
						Group:       "PluginName",
						Permissions: []ac.Permission{
							{Action: "test:action"},
							{Action: "test:action", Scope: "test:scope"},
						},
						OrgID: ac.GlobalOrgID,
					},
					Grants: []string{"Admin", "Editor"},
				},
				{
					Role: ac.RoleDTO{
						Version:     1,
						Name:        "test:name",
						Group:       "PluginName",
						Permissions: []ac.Permission{},
						OrgID:       ac.GlobalOrgID,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRegistrations("PluginName", tt.regs)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidatePluginRole(t *testing.T) {
	tests := []struct {
		name     string
		pluginID string
		role     ac.RoleDTO
		wantErr  error
	}{
		{
			name:     "empty",
			pluginID: "",
			role:     ac.RoleDTO{Name: "plugins::"},
			wantErr:  ac.ErrPluginIDRequired,
		},
		{
			name:     "invalid name",
			pluginID: "test-app",
			role:     ac.RoleDTO{Name: "test-app:reader"},
			wantErr:  &ac.ErrorInvalidRole{},
		},
		{
			name:     "invalid id in name",
			pluginID: "test-app",
			role:     ac.RoleDTO{Name: "plugins:test-app2:reader"},
			wantErr:  &ac.ErrorInvalidRole{},
		},
		{
			name:     "valid name",
			pluginID: "test-app",
			role:     ac.RoleDTO{Name: "plugins:test-app:reader"},
		},
		{
			name:     "invalid permission",
			pluginID: "test-app",
			role: ac.RoleDTO{
				Name:        "plugins:test-app:reader",
				Permissions: []ac.Permission{{Action: "invalidtest-app:read"}},
			},
			wantErr: &ac.ErrorInvalidRole{},
		},
		{
			name:     "valid permissions",
			pluginID: "test-app",
			role: ac.RoleDTO{
				Name: "plugins:test-app:reader",
				Permissions: []ac.Permission{
					{Action: "plugins.app:access"},
					{Action: "test-app:read"},
					{Action: "test-app.resources:read"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePluginRole(tt.pluginID, tt.role)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
