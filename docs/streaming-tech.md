# Streaming desktop technology summary

We want to be able to stream a desktop environment from a virtual machine on a powerful Linux host to a simple, cheap thin client such as a Raspberry Pi. Peripherals (keyboard, mouse, etc.) will be plugged into the thin client. We want to minimize latency while maintaining high quality video (1080p60 as a target). Latency should be low enough in a local environment (i.e. home gigabit ethernet) to play casual games. Video quality should be high enough in the same local environment to avoid visual artefacts in standard desktop applications (e.g. text editors or web page viewing).

In no particular order:

- [Spice](https://www.spice-space.org/features.html)
    - this is the most 'modern', yet default/standard way that virtual machines (i.e. KVM/QEMU on Linux) will share their desktop. It provides a VNC-like interface to the virtual machine, and is usually configured by default when you create a VM with `libvert`/`virt-manager`
    - This feels fine for puttering around VMs but isn't particularly fast/low latency.
    - From a Raspberry Pi (or other Linux), you can connect to your VM over the network using `remote-viewer` (from `virt-viewer`).
    - Anything requring accurate relative mouse position (e.g. games) will require USB redirection. Spice has a USB over IP tool called [`usbredir`](https://www.spice-space.org/usbredir.html) which is built in to `remote-viewer`. You only get 2 devices to forward, which is usually sufficient (keyboard/mouse)
- [Steam Link](https://store.steampowered.com/steamlink/about/)
    - intended to play games over your local network, and starts up Steam's "Big Picture" mode by default
    - [easy to install on Raspberry Pi](https://support.steampowered.com/kb_article.php?ref=6153-IFGH-6589)
    - in advanced options, you can force Steam Link to start in desktop mode, effectively making this a desktop streaming software
    - works great with an xbox controller (or similar)
    - does not have USB redirection by default - requires you to install a separate (paid) app to forward keyboard and mouse for proper gaming
    - latency feels pretty good, but can sometimes feel choppy
    - supposed to work from any platform that runs Steam. On the client side, all clients worked fine. On the hosting side, the Linux Steam Link host would crash depending on what I was trying to play. Windows host was stable.
- [Moonlight](https://moonlight-stream.org/)
    - this only works with Nvidia GPUs. It's a reverse engineering (open source) of Nvidia's GameStream technology. You also need to have GameStream enabled. This means that your virtual machine MUST have PCI passthrough on the GPU, or Moonlight will not work.
    - this only works on Windows 10 hosts. You can't install GeForce Experience on non-Windows computers.
    - seems to handle USB over IP without any fuss (no configuration necessary, your USB devices just work)
    - [easy install on Raspberry Pi](https://www.howtogeek.com/220969/turn-a-raspberry-pi-into-a-steam-machine-with-moonlight/)
    - similar latency performance to Steam Link, better video quality
    - set up to stream individual games, but can be configured to simply stream your desktop
- [Parsec](https://parsecgaming.com/)
    - often compared to Moonlight in terms of features
    - As with Moonlight:
        - must be GPU accelerated (requires PCI passthrough with VM), but works with both Nvidia and AMD cards
        - requires a Windows 10 host
        - handles USB over IP without any issues
        - easy to install Raspberry Pi client, [but does not yet work on the new Raspberry Pi 4](https://support.parsecgaming.com/hc/en-us/articles/115002699012-Setting-Up-On-Raspberry-Pi-Raspbian-)
    - Advantages over Moonlight:
        - lower latency. Parsec was significantly/visually lower latency than the other options, so I roughly benchmarked it. Using side-by-side monitors (one directly connected to the VM, one on a separate local machine using parsec) and viewing a custom flicker tool, it appeared to be about 45-60ms delayed at 1080p60, corresponding to about 3-4 frames of lag at 60Hz.
        - this latency is actually HIGHER than we should necessarily expect, because in my setup Parsec's console is reporting ~2ms decode, ~10ms encode, and ~0.5ms network latency... so there may still be some configuration optimizations/fixes to bring the latency down further
        - more consistent framerate. [Parsec touts this on their blog](https://blog.parsecgaming.com/steam-in-home-streaming-latency-test-versus-parsec-7884144b29f1), and it is fairly obvious when viewing streams - there is no stuttering or inconsistency
        - cleaner, easier to use client - obvious, front-page feature to launch directly to desktop
        - the client also makes discovering the host machines very easy: you have to log in to their service, but can then discover any/all machines you've set up to stream with Parsec. Qualitatively the smoothest experience of any of the streaming services.
- [NDI](https://www.ndi.tv/)
    - This is a video streaming technology ONLY. For all of my tests, I actually used Parsec as the method to transmit USB inputs (ignoring the video streaming part of Parsec)
    - This is meant to be an extremely low latency, proprietary protocol for video production in local networking environments
    - NDI is only easily available on Windows. However, there are plugins/paid apps/etc for Linux and OSX. There is no Raspberry Pi support yet, but with the new Raspberry Pi 4, it may be possible
    - NDI uses only CPU and not GPU for both encoding and decoding. NDI states that it should use "about 5% CPU on a modern i7 for encoding/decoding", which is optimistic. I tested playing Overwatch with and without NDI on an i7, and using NDI approximately doubled the CPU usage (e.g. from 10% to 20% CPU usage). This poses an issue for lower power devices such as Raspberry Pi (which are more suited for GPU decoding, e.g. the Raspberry Pi 4 has a dedicated h264 decoder, which is useless in this application) 
    - To test out NDI, I had to run `NDI scan converter` on a Windows VM, and `NDI studio monitor` on a client (Windows) PC. These are freely available from NDI - and NDI also has a free SDK for developing on Linux/Mac/embedded devices, although it is fairly difficult to use (companies will spend a lot of time/money to develop apps or devices for NDI decoding or encoding, and then license the SDK)
    - This also felt fairly low latency. Using the same flicker test as with Parsec, I was getting around 100-120ms of overall latency. As with Parsec, this is much higher than the theoretical limit, although it is difficult to determine where the latency is occurring as there are no readily available profiling tools
- [Chrome remote desktop](https://remotedesktop.google.com/)
    - this is similar to something like VNC, TeamViewer, or Spice.
    - no USB device forwarding (unusable for some games)
    - latency feels similar to Moonlight, Steam Link, etc.
    - requires you to run Chrome in the background 
    - some handy tools for file transfer between guest/host (not really relevant for our application)
