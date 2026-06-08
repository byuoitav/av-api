/**
Defines structs that get passed over the entire project
**/
package structs

type Building struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Shortname   string `json:"shortname,omitempty"`
	Description string `json:"description,omitempty"`
}

/*Command represents all the information needed to issue a particular command to a device.
Name: Command name
Endpoint: the endpoint within the microservice
Microservice: the location of the microservice to call to communicate with the device.
Priority: The relative priority of the command relative to other commands. Commands
					with a higher (closer to 1) priority will be issued to the devices first.
*/
type Command struct {
	Name         string   `json:"name"`
	Endpoint     Endpoint `json:"endpoint"`
	Microservice string   `json:"microservice"`
	Priority     int      `json:"priority"`
}

/*RawCommand represents all the information needed to issue a particular command to a device.
Name: Command name
Description: command description
Priority: The relative priority of the command relative to other commands. Commands
					with a higher (closer to 1) priority will be issued to the devices first.
*/
type RawCommand struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

//CommandSorterByPriority sorts commands by priority and implements sort.Interface
type CommandSorterByPriority struct {
	Commands []RawCommand
}

//Len is part of sort.Interface
func (c *CommandSorterByPriority) Len() int {
	return len(c.Commands)
}

//Swap is part of sort.Interface
func (c *CommandSorterByPriority) Swap(i, j int) {
	c.Commands[i], c.Commands[j] = c.Commands[j], c.Commands[i]
}

//Less is part of sort.Interface
func (c *CommandSorterByPriority) Less(i, j int) bool {
	return c.Commands[i].Priority < c.Commands[j].Priority
}

//RoomConfiguration reflects a defined room configuration with the commands and
//command keys incldued therein.
type RoomConfiguration struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	RoomKey     string                   `json:"roomKey"`
	Description string                   `json:"description"`
	RoomInitKey string                   `json:"roomInitKey"`
	Evaluators  []ConfigurationEvaluator `json:"evaluators"`
}

//ConfigurationEvaluator commands is the command information correlated with the
//specifics for the configuration (key and priority)
type ConfigurationEvaluator struct {
	Priority     int    `json:"priority"`
	EvaluatorKey string `json:"evaluatorKey"`
}

//Device represents a device object as found in the DB.
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

//GetFullName reutrns the string of building + room + name
func (d *Device) GetFullName() string {
	return (d.Building.Shortname + "-" + d.Room.Name + "-" + d.Name)
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

//DeviceType corresponds to the DeviceType table in the database
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

//this is a stopgap to set the attribute
//TODO: rework this
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

//DeviceType corresponds to the DeviceType table in the database
type DeviceClass struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	DisplayName string `json:"display-name"`
	Description string `json:"description"`
}

//Port represents a physical port on a device (HDMI, DP, Audio, etc.)
//TODO: this corresponds to the PortConfiguration table in the database!!!
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

//Endpoint represents a path on a microservice.
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

//PortType corresponds to the Ports table in the Database and really should be called Port
//TODO:Change struct name to "Port"
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

//Room represents a room object as represented in the DB.
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
