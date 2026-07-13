# What changed since 2020 (written 07-13-2026)

The original docs ([streaming-tech.md](streaming-tech.md), [early-devlogs.md](early-devlogs.md)) captured the 2020 landscape. Six years later, some of the hardest problems from those notes have dissolved. What got cheaper, and what didn't:

## Software (free and open now)

- In 2020, Moonlight had the best latency but required an Nvidia GPU AND a Windows 10 host running GeForce Experience (see the notes in `streaming-tech.md`)
- [Sunshine](https://github.com/LizardByte/Sunshine) has since appeared: an open-source, Moonlight-compatible host
    - runs on Linux (no more Windows-host requirement)
    - hardware encoding on Nvidia, AMD, and Intel GPUs
    - the "must be a Windows 10 host with GameStream" constraint is simply gone
- Moonlight clients now run on basically everything: Raspberry Pi, Apple TV, Android sticks, web browsers
- Parsec still exists (acquired by Unity), but the open stack has caught up
- USB/keyboard/mouse forwarding is handled by these tools out of the box — the entire `usbip` rabbit hole from the early devlogs is no longer necessary

## Thin clients per dollar

- used corporate mini PCs (Lenovo ThinkCentre Tiny, Dell OptiPlex Micro, HP EliteDesk Mini) flooded the second-hand market: $40–80 for a real x86 CPU with hardware H.264/HEVC decode — embarrasses a Pi 4 as a streaming client
- a ~$30 Android TV stick running Moonlight also beats the 2020 Pi experience
- the Pi itself is the exception: the 2021–2023 shortage had $35 boards scalping for $100+, and the Pi 5 nominally costs more than the Pi 4 did

## Video encoding hardware is everywhere

- in 2020, NDI's CPU-only encoding doubling CPU usage was a real finding
- every GPU of the last few generations now has excellent low-latency hardware encoders, including AV1
- an Intel Arc A310 (~$100) does hardware AV1 encode
- the Pi 5 dropping the hardware H.264 *encoder* is irrelevant for this project's direction of streaming — the client only decodes

## Networking

- 2.5GbE went from exotic to bundled on cheap motherboards; used 10GbE gear is thrift-store priced
- the raw uncompressed approach (see `src/stream`: ~280 MB/s for 1470x956@50 RGBA) saturates gigabit but fits comfortably on a 2.5GbE link — uncompressed, zero-codec-latency streaming for ~$20 in NICs is now plausible on a home network

## What got MORE expensive

- new GPUs at the high end
- RAM: DRAM prices spiked hard in late 2025 on AI demand, so a big-RAM VM host costs more to build today than in 2020

## Implication for bigboy

The thing this project was building toward — low-latency Linux-host desktop streaming to a cheap client — went from "research project with USB/IP rabbit holes" to `apt install sunshine` plus a $50 used mini PC. Which arguably makes the homebrew version (`src/stream`) more fun, since nothing's at stake anymore.
