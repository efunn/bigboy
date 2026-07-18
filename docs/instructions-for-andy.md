# Server-side setup (Texas) — for Andy

You're setting up the host end of a Sunshine → Moonlight stream. Client is in the
Yukon on Northwestel fiber: nominally "15 Mbit" (really ~20), **200 GB/month cap**,
almost certainly behind CGNAT. Full context and the client-side config are in
[remote-streaming-setup.md](remote-streaming-setup.md); this doc is just the Texas box.

Machine: Windows, GTX 1080, gigabit up. The 1080's 6th-gen NVENC does H.264 and
HEVC in hardware — we want **HEVC**, since the client is bandwidth-starved and
HEVC buys ~30–50% quality-per-bit over H.264 at these bitrates.

## 1. Sunshine

- Install Sunshine (the LizardByte build) on Windows. Run it as a service so it
  survives logout / is up before login.
- Web UI is at https://localhost:47990 — set the admin credentials on first launch.
  **Keep 47990 local; never forward it.**
- Encoder settings (Configuration → Video):
  - Encoder: **NVENC**
  - Codec: allow HEVC (client will request it)
  - NVENC preset: favor quality over speed — encode time is trivial next to the
    Texas↔Yukon RTT, so there's no reason to run a low-latency/low-quality preset.
    P5–P6 range is fine.
- Bitrate and resolution are **driven by the client** (Moonlight), not set here, so
  you don't need to pick them — but see FEC below, which is yours to tune.
- **Lossy-WAN tuning:** bump the FEC / redundancy percentage above the default.
  This trades a bit of bandwidth for resilience to packet loss on the long haul,
  which matters far more over 3,000 km than on a LAN. Start moderate and let the
  client report loss.

## 2. Keep it reachable and awake (Windows gotchas)

- Disable sleep and "fast startup"; set the power plan so the box stays up. A
  sleeping host is the #1 "why can't I connect" cause.
- If the GPU driver enters low-power / no-display state with no monitor attached,
  NVENC can misbehave. A cheap **HDMI dummy plug** (headless display emulator)
  keeps a virtual display alive and lets the client pick real resolutions. Worth
  having one in the machine.
- Give the box a static LAN IP or a DHCP reservation — you'll want it stable for
  either networking option below.

## 3. Networking — pick one

### Option A: Tailscale (recommended, and probably required)

The Yukon side is very likely CGNAT'd, and Texas may be too. Tailscale (WireGuard)
sidesteps both and needs no router config.

- Install Tailscale on this box, sign in, join the tailnet the client will use.
- Moonlight on the client connects to **this machine's Tailscale IP** (100.x.y.z).
- Verify you get a **direct** connection rather than a relayed one:
  - `tailscale ping <client-hostname>` — it'll report "direct" or "via DERP …".
  - `tailscale netcheck` — shows your NAT type / UDP situation.
  - DERP-relayed works but adds latency; if it won't go direct, check that UDP
    isn't being blocked and consider enabling Tailscale's direct-connection
    prerequisites (both ends need outbound UDP). For a fixed pair, DERP is usually
    only a mild penalty.
- Optional: assign this host a stable MagicDNS name so the client isn't chasing IPs.

### Option B: Direct port forwarding (lowest latency, only if Texas has a real public IP)

**First check for CGNAT.** Compare the WAN IP shown in your router's status page to
the public IP from whatismyip.com (or `curl ifconfig.me` from the host). If they
differ — or the router WAN IP is in 100.64.0.0/10 — you're behind CGNAT and Option B
won't work; use Tailscale. If they match, you have a routable IP and can forward:

- Forward to this host's LAN IP:
  - **TCP:** 47984, 47989, 48010
  - **UDP:** 47998, 47999, 48000, 48010
- Do **not** forward 47990 (web UI).
- Allow those same ports through Windows Defender Firewall (the Sunshine installer
  usually adds rules; verify).
- Prefer explicit forwards over UPnP; disable UPnP on the router if you can.
- Client points Moonlight at your public IP (or a dynamic-DNS name if the IP isn't
  static — Northwestel/residential IPs usually aren't).

You can also run **both**: forward ports *and* have Tailscale as a fallback path.

## 4. Pairing

1. Client adds the host in Moonlight (Tailscale IP, or public IP if forwarding).
2. Sunshine shows a PIN prompt (or the client enters a PIN in Sunshine's web UI
   under "PIN"). Enter it once to pair; it persists.

## 5. Sanity checks before handing off

- From the client, confirm Moonlight lists the host and pairs.
- Start a session at a low mode first (e.g. 720p60, ~6 Mbit HEVC) and watch
  Moonlight's on-screen stats (Ctrl+Alt+Shift+S in the PC client) for loss,
  decode time, and host encode time.
- Encode time should be a couple ms; if network loss is high, raise FEC (§1) and/or
  have the client drop bitrate.
- Remind the client that the **200 GB cap**, not the link speed, is the real limit
  — server side you don't control bitrate, so just make sure HEVC is actually the
  negotiated codec (visible in the Moonlight stats overlay).
