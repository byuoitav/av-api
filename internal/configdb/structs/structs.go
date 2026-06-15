package structs

type Building struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Shortname   string `json:"shortname,omitempty"`
	Description string `json:"description,omitempty"`
}

type Command struct {
	Name         string   `json:"name"`
	Endpoint     Endpoint `json:"endpoint"`
	Microservice string   `json:"microservice"`
	Priority     int      `json:"priority"`
}

type RawCommand struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

type CommandSorterByPriority struct {
	Commands []RawCommand
}

func (c *CommandSorterByPriority) Len() int {
	return len(c.Commands)
}

func (c *CommandSorterByPriority) Swap(i, j int) {
	c.Commands[i], c.Commands[j] = c.Commands[j], c.Commands[i]
}

func (c *CommandSorterByPriority) Less(i, j int) bool {
	return c.Commands[i].Priority < c.Commands[j].Priority
}

type RoomConfiguration struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	RoomKey     string                   `json:"roomKey"`
	Description string                   `json:"description"`
	RoomInitKey string                   `json:"roomInitKey"`
	Evaluators  []ConfigurationEvaluator `json:"evaluators"`
}

type ConfigurationEvaluator struct {
	Priority     int    `json:"priority"`
	EvaluatorKey string `json:"evaluatorKey"`
}

type Device struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Address     string    `json:"address"`
	Input       bool      `json:"input"`
	Output      bool      `json:"output"`
	Building    Building  `json:"building"`
	Room        Room      `json:"room"`
	Type        string    `json:"type"`
	Class       string    `json:"class,omitempty"`
	Power       string    `json:"power"`
	Roles       []string  `json:"roles,omitempty"`
	Blanked     *bool     `json:"blanked,omitempty"`
	Volume      *int      `json:"volume,omitempty"`
	Muted       *bool     `json:"muted,omitempty"`
	PowerStates []string  `json:"powerstates,omitempty"`
	Responding  bool      `json:"responding"`
	Ports       []Port    `json:"ports,omitempty"`
	Commands    []Command `json:"commands,omitempty"`
}

func (d *Device) GetFullName() string {
	return d.Building.Shortname + "-" + d.Room.Name + "-" + d.Name
}

func (p *Device) HasRole(r string) bool {
	for _, role := range p.Roles {
		if r == role {
			return true
		}
	}

	return false
}

func HasRole(d Device, r string) bool {
	for _, role := range d.Roles {
		if r == role {
			return true
		}
	}

	return false
}

func RoleId(device Device, roleId int) bool {
	return false
}

func (p *Device) GetCommandByName(commandName string) Command {
	for _, command := range p.Commands {
		if command.Name == commandName {
			return command
		}
	}

	return Command{}
}

type DeviceCommand struct {
	ID             int  `json:"id,omitempty"`
	DeviceID       int  `json:"device"`
	CommandID      int  `json:"command"`
	MicroserviceID int  `json:"microservice"`
	EndpointID     int  `json:"endpoint"`
	Enabled        bool `json:"enabled"`
}

type DeviceType struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DevicePowerState struct {
	ID           int `json:"id,omitempty"`
	DeviceID     int `json:"device"`
	PowerStateID int `json:"powerstate"`
}

type DeviceAttributeInfo struct {
	DeviceID       int    `json:"deviceID"`
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
	AttributeType  string `json:"attributeType"`
}

type DeviceRole struct {
	ID                     int `json:"id,omitempty"`
	DeviceID               int `json:"device"`
	DeviceRoleDefinitionID int `json:"role"`
}

type DeviceRoleDef struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DeviceClass struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	DisplayName string `json:"display-name"`
	Description string `json:"description"`
}

type Port struct {
	Source      string `json:"source"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
	Host        string `json:"host"`
}

type PortConfiguration struct {
	ID                  int `json:"id,omitempty"`
	DestinationDeviceID int `json:"destination-device"`
	PortID              int `json:"port"`
	SourceDeviceID      int `json:"source-device"`
	HostDeviceID        int `json:"host-device"`
}

type Endpoint struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type Microservice struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type PortType struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DeviceTypePort struct {
	DeviceTypePortID     int      `json:"id"`
	DeviceTypeID         int      `json:"type-id"`
	DeviceTypeName       string   `json:"type-name"`
	Port                 PortType `json:"port-info"`
	Description          string   `json:"type-port-description"`
	FriendlyName         string   `json:"friendlyName"`
	HostDestintionMirror bool     `json:"mirror-host-dest"`
}

type PowerState struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Room struct {
	ID              int               `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Description     string            `json:"description,omitempty"`
	Building        Building          `json:"building,omitempty"`
	Devices         []Device          `json:"devices,omitempty"`
	ConfigurationID int               `json:"configurationID,omitempty"`
	Configuration   RoomConfiguration `json:"configuration"`
	RoomDesignation string            `json:"roomDesignation"`
}
