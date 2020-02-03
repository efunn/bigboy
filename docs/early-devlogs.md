# Log of activities

## 01-09-2020

- assembled computer and reinstalled OS (Ubuntu 18.0.4 LTS)
- raspberry pi powers on, needs mini-HDMI for graphics to test further
- tested USB over IP on localhost
    - using `usbip` package installed from `linux-tools-generic` (NOT standalone)
    - requires specific/matching kernel version of package (e.g. `linux-tools-5.0.0-37-generic`)
    - `usbip list` gives a generic error (`cannot find /path/to/missing/file`) that can be avoided by creating dummy file in location (doesn't really make a difference though)
    - identify USB bus of devices with the device's ID (e.g. `05ac:024f`) using `usbip list -l | grep 'busid' | grep '05ac:024f' | cut '-d ' -f4`
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

## 01-10-2020

- backgroud reading on KVM, libvert, QEMU, VFIO
    - KVM/QEMU used to be from the same project - now KVM is the generic kernel virtualization (Type-1 hypervisor), and QEMU is by itself a Type-2 hypervisor, but with KVM enabled with QEMU, it effectively becomes a Type-1 hypervisor
    - libvert is a daemon and set of CLI tools to manage VMs, we'll be using KVM/QEMU here (libvert can do the stuff that you might use something like virtualbox to manage)
    - some graphical interfaces for libvert as well (you can also use something like AQEMU for 'direct' management of QEMU)
    - other notes from Andy: LXD on Ubuntu may offer better performance; can potentially access LXD with VNC (simple setup)
- acquired micro-HDMI cable for raspberry pi
- raspberry pi works with small display; but couldn't adjust `config.txt` in `/boot` to work with ultrawide display
- `usbip` not in standard packages on Raspbian for correct kernel version; would need to build separately (weird rabbit hole to dive into, so would like to avoid)
- trying to get Ubuntu for raspberry pi working
    - running into power issues: would only boot for a few seconds before crashing; unplugging fan (source of power draw) would let it boot slightly longer
    - tried various power sources; plugging USB-C directly from big computer motherboard allowed it to boot properly (albeit with periodic `Under-voltage detected!` errors and a yellow lightning bolt warning on the screen)
    - proper power supply ordered (known issue with raspberry pi 4)
    - installing `lubuntu` GUI to play around with
- `usbip` still failing on separate machines (with raspberry pi as host) - uninformative errors, might just be broken in general (many people reporting similar issues with no resolution)
- next steps:
    - explore other options for USB redirection - set up spice on VM and raspberry pi


## week ending with 01-17-2020

- tested various streaming protocols this week
- `spice` works poorly for video/games, well for desktop; with USB devices properly forwarded, keyboard/mouse inputs are sensible and relatively quick; without USB forwarding games are practically unplayable with mouse 
- streaming youtube from `spice` on powerful computer was actually better in quality than running youtube locally on the raspberry pi
- VLC/youtube players are not hardware accelerated on raspberry pi; e.g. playing a 4k sample clip on VLC is choppy, but using `omxplayer` (which uses the `h264` decoder on the rpi4) runs smoothly
- tried streaming UDP from OBS to `omxplayer` on raspberry pi; dropped packets frequently, but quality would be okay (significantly delayed - approx 6 seconds); packet loss led to very odd artefacts
- tested steamlink: works very well from windows or mac to rpi4, although certain actions (alt-tabbing, etc.) could sometimes lock everything up; steamlink from ubuntu would hang after 3 seconds (probably some sort of graphics issue from ubuntu; most suggestions online related to changing video settings or installing drivers; did not find a working configuration, though)
- tried to test NDI locally on powerful computer (no easy NDI monitor for raspberry pi/arm): found a project with a simple monitor, but was only built for the NDI 3.8 SDK, while you can only download the NDI 4.1 SDK now; there is a unity plugin built for NDI 4.X that we will try next
- next steps:
    - test NDI locally on ubuntu (find linux viewer for NDI 4.X, or build unity program separately)
    - test other, promising streaming protocols: moonlight (uses reverse-engineered nvidia gamestream, requires nvidia card, but code is open source), or parsec (works on any platform, free, but code not available)
    - parsec not yet working on rpi4, but compatible with rpi<=3; will test with raspberry pi 3 to check performance
