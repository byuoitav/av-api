Logical flow of room state change in the AV-API

### Put definition ###

Put to the rooms endpoint, which has the *building*, and *room* as URL parameters
and contains a payload containing a JSON document with one or more of the following
options:
  * *CurrentVideoInput*: The desired video input (source) for all video output devices
  in the room.
  * *CurrentAudioInput*: The desired  audio input (source) for all audio output devices
  in the room.
  * *Displays*: An array of display (video output) devices with desired state defined
  in the provided values. Potential properties:
    * *Name*: **Required** The name of the device. (D1, CP1, roku, etc.) This can
    also be the device type (TV, Projector, etc.)
    * *Power*: The desired power state of the device.
    * *Blanked*: true/false blank the display.
    * *Input*: The desired video input for the device. This will override the room
    defined currentVideoInput.
  * *AudioDevices*: an array of audio devices (Audio output) with desired state defined
  in the provided values. Potential properties:
    * *Name*: **Required** The name of the device. (D1, CP1, roku, etc.)
    * *Power*: The desired power state of the device.
    * *Input*: The desired video input for the device. This will override the room
    defined currentVideoInput.
    * *Muted*: true/false mute the audio.
    * *volume*: numeric value defining the desired percentage of volume output (0-100)


  Note that a device can exist both in the Displays and the AudioDevices arrays,
  as those are *logical* distinctions and there are devices that fill both Logical
  roles (e.g. a television)

### Logical Flow of API ###
