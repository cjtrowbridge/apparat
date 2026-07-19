package hud

type TabPlacement string

const (
	TabPlacementTop        TabPlacement = "top"
	TabPlacementLeftRail   TabPlacement = "left_rail"
	TabPlacementRightRail  TabPlacement = "right_rail"
	TabPlacementResponsive TabPlacement = "responsive"
)

type TabDensity string

const (
	TabDensityComfortable TabDensity = "comfortable"
	TabDensityCompact     TabDensity = "compact"
	TabDensityExpanded    TabDensity = "expanded"
)

type TabLabelMode string

const (
	TabLabelIconText TabLabelMode = "icon_text"
	TabLabelIconOnly TabLabelMode = "icon_only"
	TabLabelTextOnly TabLabelMode = "text_only"
)

type Theme string

const (
	ThemeDark         Theme = "dark"
	ThemeLight        Theme = "light"
	ThemeHighContrast Theme = "high_contrast"
)

type HUDConfig struct {
	TabView       TabViewConfig
	Bindings      BindingConfig
	Display       DisplayConfig
	Interaction   InteractionConfig
	Notifications NotificationConfig
	Diagnostics   DiagnosticConfig
	DefaultViews  DefaultViewConfig
	Privacy       PrivacyConfig
}

type TabViewConfig struct {
	Placement    TabPlacement
	Allowed      []TabPlacement
	Density      TabDensity
	LabelMode    TabLabelMode
	DefaultTab   TabID
	Tabs         []TabDescriptor
	RememberLast bool
}

type TabDescriptor struct {
	ID                 TabID
	Label              string
	Glyph              string
	AccessibilityLabel string
	Visible            bool
	Enabled            bool
	Badge              string
	Status             string
}

type BindingConfig struct {
	Bindings map[Action][]Binding
}

type Binding struct {
	Device string
	Input  string
}

type DisplayConfig struct {
	Theme                 Theme
	AccentColor           string
	Scale                 float64
	FontSize              int
	FontFamily            string
	ButtonBackgroundColor string
	PanelBackgroundColor  string
	ReducedMotion         bool
	Contrast              string
	FocusRingStrength     string
	PanelDensity          string
	ListRowHeight         int
	CardSpacing           int
	TextOverflow          string
}

type InteractionConfig struct {
	ControllerRepeatDelayMS int
	KeyboardRepeatDelayMS   int
	MouseScrollSpeed        float64
	PushToTalkMode          string
	ConfirmDestructive      bool
	CommandPaletteVisible   bool
	CommandPaletteShortcut  string
	LandingTab              TabID
	RememberSelections      bool
}

type NotificationConfig struct {
	Visibility  string
	Muted       bool
	Volume      float64
	ToastMS     int
	FocusUrgent bool
	Categories  []string
	QuietHours  string
}

type DiagnosticConfig struct {
	DeveloperOverlay bool
	LogDetail        string
	ShowRuntimePaths bool
	ShowBuildPaths   bool
	ShowFrameTiming  bool
	ShowMemory       bool
	ShowInputEvents  bool
	ShowFocusPath    bool
	ShowLayoutBounds bool
}

type DefaultViewConfig struct {
	Projects string
	Cluster  string
	Tasks    string
	Comrades string
	Research string
}

type PrivacyConfig struct {
	HideSensitivePresentation bool
	RequireRevealForSecrets   bool
	SharingDefault            string
}

type ConfigManager interface {
	Config() HUDConfig
}

type DefaultConfigManager struct{}

func (DefaultConfigManager) Config() HUDConfig {
	tabs := DefaultTabDescriptors()
	return HUDConfig{
		TabView: TabViewConfig{
			Placement:  TabPlacementTop,
			Allowed:    []TabPlacement{TabPlacementTop, TabPlacementLeftRail, TabPlacementRightRail, TabPlacementResponsive},
			Density:    TabDensityComfortable,
			LabelMode:  TabLabelIconText,
			DefaultTab: TabComrades,
			Tabs:       tabs,
		},
		Bindings:      DefaultBindings(),
		Display:       DisplayConfig{Theme: ThemeDark, AccentColor: "apparat-blue", Scale: 1.0, FontSize: 18, FontFamily: "MPlus1pRegular", ButtonBackgroundColor: "accent_color", PanelBackgroundColor: "panel_color", Contrast: "normal", FocusRingStrength: "strong", PanelDensity: "comfortable", ListRowHeight: 32, CardSpacing: 12, TextOverflow: "wrap"},
		Interaction:   InteractionConfig{ControllerRepeatDelayMS: 180, KeyboardRepeatDelayMS: 140, MouseScrollSpeed: 1.0, PushToTalkMode: "hold", ConfirmDestructive: true, CommandPaletteVisible: true, CommandPaletteShortcut: "Ctrl+Shift+P", LandingTab: TabComrades},
		Notifications: NotificationConfig{Visibility: "important_local_events", Volume: 0.5, ToastMS: 5000, Categories: []string{"job_completion", "device_online", "device_offline", "task_failure", "comrade_request", "research_milestone", "security_warning"}, QuietHours: "not_configured"},
		Diagnostics:   DiagnosticConfig{LogDetail: "info", ShowRuntimePaths: true, ShowBuildPaths: true, ShowFrameTiming: true, ShowMemory: true, ShowInputEvents: true, ShowFocusPath: true, ShowLayoutBounds: true},
		DefaultViews:  DefaultViewConfig{Projects: "recent_projects", Cluster: "device_health_summary", Tasks: "active_and_failed_runs", Comrades: "placeholder_relationships", Research: "validated_project_catalog"},
		Privacy:       PrivacyConfig{HideSensitivePresentation: true, RequireRevealForSecrets: true, SharingDefault: "disabled"},
	}
}
