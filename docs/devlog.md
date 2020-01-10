# Daily log of activities

## 01-09-2020

- assembled computer and reinstalled OS (Ubuntu 18.0.4 LTS)
- raspberry pi powers on, needs mini-HDMI for graphics to test further
- tested USB over IP on localhost
    - using `usbip` package installed from `linux-tools-generic` (NOT standalone)
    - requires specific/matching kernel version of package (e.g. `linux-tools-5.0.0-37-generic`)
    - `usbip list` gives a generic error (`cannot find /path/to/missing/file`) that can be avoided by creating dummy file in location (doesn't really make a difference though)
    - identify USB bus of devices with the device's ID (e.g. `05ac:024f`) using `usbip list -l | grep 'busid' | grep '05ac:024f` | cut '-d ' -f4`
    - should load appropriate kernel modules with `modprobe` or it won't work
    - on 'server', run `usbip bind --busid=5-2` (busid found from previous `grep`)
    - on 'client', run `usbip attach -r localhost -b 5-2`
    - this currently tries to connect but fails, possibly because both server and client are running on localhost (device can't be on both? or some other failsafe?)
    - this should all be automated/cleaned up, services autostart, etc.
- discussed next steps for streaming video
    - would like to test SPICE client on raspberry pi
    - explore possibility of NDI streaming - no current client on raspberry pi, but can test latency on equivalent modern hardware
- next steps
    - set up `usbip` on raspberry pi
    - test if `bind` and `attach` work with different physical machines
    - set up `libvert` with `KVM` backend on host computer (most similar to unraid setup)
    - if `bind` and `attach` working, test if a virtual machine can access the virtual USB device
    - if this fails, test if `bind` and `attach` can work if the virtual machine gains access to the virtual USB device directly
    - run NDI host and test latency over local network (e.g. from OBS on one computer in the same room)
    - set up SPICE client on raspberry pi

## 02-09-2020

t.b.d.

