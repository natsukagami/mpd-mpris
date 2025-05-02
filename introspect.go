package mpris

import (
	introspect "github.com/godbus/dbus/v5/introspect"
)

// IntrospectNode returns the root node of the library's introspection output.
func (i *Instance) IntrospectNode() *introspect.Node {
	return &introspect.Node{
		Name: i.Name(),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			introspect.Interface{
				Name: "org.mpris.MediaPlayer2",
				Properties: []introspect.Property{
					introspect.Property{
						Name:   "CanQuit",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanRaise",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "HasTrackList",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "Identity",
						Type:   "s",
						Access: "read",
					},
					introspect.Property{
						Name:   "SupportedUriSchemes",
						Type:   "as",
						Access: "read",
					},
					introspect.Property{
						Name:   "SupportedMimeTypes",
						Type:   "as",
						Access: "read",
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Raise",
					},
					introspect.Method{
						Name: "Quit",
					},
				},
			},
			introspect.Interface{
				Name: "org.mpris.MediaPlayer2.Player",
				Properties: []introspect.Property{
					introspect.Property{
						Name:   "PlaybackStatus",
						Type:   "s",
						Access: "read",
					},
					introspect.Property{
						Name:   "LoopStatus",
						Type:   "s",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Rate",
						Type:   "d",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Shuffle",
						Type:   "b",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Metadata",
						Type:   "a{sv}",
						Access: "read",
					},
					introspect.Property{
						Name:   "Volume",
						Type:   "d",
						Access: "readwrite",
					},
					introspect.Property{
						Name:   "Position",
						Type:   "x",
						Access: "read",
					},
					introspect.Property{
						Name:   "MinimumRate",
						Type:   "d",
						Access: "read",
					},
					introspect.Property{
						Name:   "MaximumRate",
						Type:   "d",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanGoNext",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanGoPrevious",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanPlay",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanSeek",
						Type:   "b",
						Access: "read",
					},
					introspect.Property{
						Name:   "CanControl",
						Type:   "b",
						Access: "read",
					},
				},
				Signals: []introspect.Signal{
					introspect.Signal{
						Name: "Seeked",
						Args: []introspect.Arg{
							introspect.Arg{
								Name: "Position",
								Type: "x",
							},
						},
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Next",
					},
					introspect.Method{
						Name: "Previous",
					},
					introspect.Method{
						Name: "Pause",
					},
					introspect.Method{
						Name: "PlayPause",
					},
					introspect.Method{
						Name: "Stop",
					},
					introspect.Method{
						Name: "Play",
					},
					introspect.Method{
						Name: "Seek",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "Offset",
								Type:      "x",
								Direction: "in",
							},
						},
					},
					introspect.Method{
						Name: "SetPosition",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "TrackId",
								Type:      "o",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "Position",
								Type:      "x",
								Direction: "in",
							},
						},
					},
				},
			},
			introspect.Interface{
				Name: "org.freedesktop.DBus.Properties",
				Signals: []introspect.Signal{
					introspect.Signal{
						Name: "PropertiesChanged",
						Args: []introspect.Arg{
							introspect.Arg{
								Name: "interface_name",
								Type: "s",
							},
							introspect.Arg{
								Name: "changed_properties",
								Type: "a{sv}",
							},
						},
					},
				},
				Methods: []introspect.Method{
					introspect.Method{
						Name: "Get",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "property_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "value",
								Type:      "v",
								Direction: "out",
							},
						},
					},
					introspect.Method{
						Name: "GetAll",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "properties",
								Type:      "a{sv}",
								Direction: "out",
							},
						},
					},
					introspect.Method{
						Name: "Set",
						Args: []introspect.Arg{
							introspect.Arg{
								Name:      "interface_name",
								Type:      "s",
								Direction: "in",
							},
							introspect.Arg{
								Name:      "property_name",
								Type:      "s",
								Direction: "out",
							},
							introspect.Arg{
								Name:      "value",
								Type:      "v",
								Direction: "in",
							},
						},
					},
				},
			},
			// TODO: This interface is not fully implemented.
			// introspect.Interface{
			// 	Name: "org.mpris.MediaPlayer2.TrackList",

			// },
		},
	}
}
