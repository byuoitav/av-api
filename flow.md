## **tl;dr** ##
The AV API makes decisions based on a user-provided JSON payload and information in a database to set devices to a desired state.

## PUT definition ##

PUT to the rooms endpoint, which has the *building*, and *room* as URL parameters
and contains a payload containing a JSON document with one or more of the following
options:
  * *CurrentVideoInput*: The desired video input (source) for all video output devices
  in the room.
  * *CurrentAudioInput*: The desired  audio input (source) for all audio output devices
  in the room.
  * *Displays*: An array of display (video output) devices with desired state defined in the provided values. Potential properties:
    * *Name*: **Required** The name of the device. (D1, CP1, roku, etc.) This can
    also be the device type (TV, Projector, etc.)
    * *Power*: The desired power state of the device.
    * *Blanked*: true/false blank the display.
    * *Input*: The desired video input for the device. This will override the room
    defined currentVideoInput.
  * *AudioDevices*: an array of audio devices (Audio output) with desired state defined in the provided values. Potential properties:
    * *Name*: **Required** The name of the device. (D1, CP1, roku, etc.) This can
    also be the device type (TV, Projector, etc.)
    * *Power*: The desired power state of the device.
    * *Input*: The desired video input for the device. This will override the room
    defined currentVideoInput.
    * *Muted*: true/false mute the audio.
    * *volume*: numeric value defining the desired percentage of volume output (0-100)


  Note that a device can exist both in the Displays and the AudioDevices arrays,
  as those are *logical* distinctions and there are devices that fill both Logical
  roles (e.g. a television)

## Logical Flow of API ##

### Overview ###
Essentially there are 6 properties of rooms that can be changed via the API, they are:

1. Video Input
1. Audio Input
1. Display Blanked
1. Volume
1. Device Muted
1. Power

It is important to note that we can set the Audio/Video input on a room wide basis, but these are functionally aliases to set Audio and Video input for each
individual AudioOut and VideoOut device in the room.

The API checks the put body for which attributes were passed in, upon presence of an attribute the API runs through the logic defined below to set the properties of the devices to the desired state.

The mapping of put values to properties is included for the sake of completeness:

|PutBodyField|Property|
|---|---|
|CurrentVideoInput  |  Video Input (for all devices with the *videoOut* role in room)|
|CurrentAudioInput  |  Audio Input (for all devices with the *audioOut* role in room)|
|Display/Power      |  Power|
|Display/Input      |  VideoInput|
|Display/Blanked    |  Display Blanked|
|AudoDevice/Power   |  Power|
|AudioDevice/Volume |  Volume|
|AudioDevice/Muted  |  DeviceMuted |


### Logical Flow per Property ###

#### Video Input ####

###### Payload Requirements ######
It is important to note that in addition the the fields defined in the payload the building and room are defined by the user via URL parameter.

The required payload to change the video input at minimum:

```
{
  "currentVideoInput": "NameHere"
}
```

However if more granular control is required the post body can define inputs
on a display by display basis.

```
{
  "displays" :[{
    "name": "device1",
    "input": "inputDevice1"
    }]
}
```

The two may be mixed, with the `input` defined in the devices array overriding the `currentVideoInput` field for the devices defined. So:

```
{
  "currentVideoInput": "Input1",
  "displays" :[{
    "name": "display3",
    "input": "input2"
    }]
}
```

Would result in `display3` being set to `input2`, with *all other displays* being set to input1.

###### Logical code flow ######

Following the determination that a Video Input state has been set for one or more output devices the API needs to validate the requested state and issue a command to the devices specified to set the state. As devices traditionally set their input based on physical ports (hdmi1, hdmi2, HDBaseT, etc.) rather than what device is plugged into that port (computer, BlueRay, Roku, etc.) the API needs to retrieve the following information from the database.

**What command to issue**
* DeviceIP
* Port for the device to change input to

**Where to issue the command to**
* Microservice to communicate with the devices
* Endpoint on the relevant microservice to call

The steps vary slightly between setting an input for individual devices and the room as a whole, and both flows of logic will be represented here.

**Set video input for individual device**
For each display specified in the displays array
1. Validate that the name specified corresponds to a device in the room with the `videoOut` role.
1. Validate that the input specified corresponds to the name/type of a device in the room with the `videoIn` role.
1. Retrieve display information. The relevant information for this command will be:
  * IPAddress
  * Port Configuration (What devices are plugged into which port)
1. Validate that the input specified is plugged into the display. Retrieve the port name.  
1. Retrieve the information associated with the `ChangeInput` command for the device. Included in this will be the following information:
  * Microservice Address
  * Microservice Endpoint
    * By convention the endpoint associated with the `ChangeInput` command will have 2 parameters: `Address` and `Port`.
  * If the command is enabled.
1. Validate that the command issenabled.  
1. Replace the `Address` and `Port` URL parameters in the endpoint with the values retrieved earlier.
1. Send a GET request to the Microservice + Endpoint.
1. Confirm command success

**Set video input for all devices**
1. Validate that input specified in currentViedoInput corresponds to the name/type of a device in the room with the `videoIn` role.
1. Retrieve all devices and their information in the room with the `videoOut` role.
1. For each device specified go through steps 3-9 specified above.
