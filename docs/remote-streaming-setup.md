# Remote streaming setup: Texas → Yukon (written 07-13-2026)

Streaming a gaming PC in Texas (gigabit up, GTX 1080, Windows) to a client in
the Yukon on Northwestel fiber (cheapest tier, nominally "15 Mbit" — measures
closer to ~20, upgradeable for more money with no hardware change; 200 GB/month
cap).

This is a different problem from the LAN research in [streaming-tech.md](streaming-tech.md).
Over a 15–20 Mbit WAN link, the raw-RGBA approach in `src/stream` (~280 MB/s) is
off by ~150x and irrelevant here. This needs a hardware codec, which the 1080
provides.

## The setup

- **Server (Texas): [Sunshine](https://github.com/LizardByte/Sunshine) on Windows.**
  Runs great on Windows — it's the most common Sunshine host OS, no Linux/VM work
  needed. The GTX 1080's 6th-gen NVENC encodes H.264 and HEVC in hardware at
  near-zero CPU cost.
- **Client (Yukon): [Moonlight](https://moonlight-stream.org/).** Available for
  basically everything (PC, Mac, Android, Apple TV, browser).
- **Transport: [Tailscale](https://tailscale.com/) on both machines.** Avoids
  the port-forwarding hassle entirely — both boxes join the tailnet and Moonlight
  connects to the server's Tailscale IP. Also sidesteps CGNAT, which Northwestel
  and most northern ISPs use. See [Ports / NAT](#ports--nat) below for the
  forward-it-yourself alternative.

## The constraint that actually bites: the 200 GB cap

The 15–20 Mbit link is not the real limiter — the monthly data cap is. Cap the
Moonlight bitrate well below the link (leave ~30% headroom for loss/retransmits),
then budget the month like data, not bandwidth.

At a given bitrate (GB/hr ≈ Mbit × 0.45):

| Bitrate  | Data rate  | Hours/month at 200 GB |
|----------|------------|-----------------------|
| 4 Mbit   | 1.8 GB/hr  | ~111 hrs              |
| 6 Mbit   | 2.7 GB/hr  | ~74 hrs               |
| 8 Mbit   | 3.6 GB/hr  | ~55 hrs               |
| 10 Mbit  | 4.5 GB/hr  | ~44 hrs               |
| 12 Mbit  | 5.4 GB/hr  | ~37 hrs               |

For daily use, **8 Mbit (~55 hrs/month, ~1.8 hrs/day)** is a sane default. Since
the plan is upgradeable with no hardware change, the lever if this feels tight is
a plan bump — but check whether a faster tier also raises the 200 GB cap, since
the cap is what's rationing you, not the speed.

## Codec and resolution

- **Use HEVC (H.265), not H.264.** The 1080 encodes it in hardware, and at these
  low bitrates HEVC looks dramatically better — roughly 30–50% more quality per
  bit, exactly what you're short on. Set it in Moonlight (Settings → Video codec →
  HEVC). Any client hardware from the last several years decodes it.
- No AV1 — that needs a 40-series card. Not a factor here.

Suggested HEVC bitrates for each resolution (you said you're down to test low —
here's the ladder):

| Mode      | Suggested bitrate | Notes                                          |
|-----------|-------------------|------------------------------------------------|
| 1080p60   | 8–12 Mbit         | Aggressive; motion may smear at the low end    |
| 1080p30   | 6–8 Mbit          | Cleaner stills; good for desktop / slow games  |
| 720p60    | 5–8 Mbit          | Sharp + smooth; good all-rounder for fast games|
| 720p30    | 3–5 Mbit          | Very economical                                |
| 480p60    | 2–4 Mbit          | Ugly but responsive; stretches the data cap far|

Given the cap, **720p60 @ 6 Mbit HEVC** is probably the sweet spot: smooth,
sharp enough, and ~74 hrs/month. Bump to 1080p when the extra data is worth it.

## Switching profiles (your "on the fly" question)

Not truly live/mid-frame — Moonlight sets resolution, FPS, and bitrate *before*
each session, in the client's settings. But changing them is trivial: disconnect,
move the sliders, reconnect. Reconnect takes ~2–5 seconds, so it's effectively
instant for A/B testing. There is **no in-session hotkey to change resolution or
bitrate.** To rapidly compare 1080p60 / 720p60 / 720p30 / 480p, just adjust and
relaunch between each — ideal for finding where quality-vs-cap lands for you.

(The bitrate is a single slider; resolution and FPS are pickers. All three live
in Moonlight's Settings on the client, so the whole loop is client-side.)

## Latency

Good news from the fiber detail: Northwestel fiber means the local loop is low
latency and low jitter — the earlier satellite-backhaul worst case doesn't apply.
What remains is pure distance: Texas → Whitehorse is ~3,000+ km, and the route
likely backhauls through Vancouver/Seattle, so expect ~50–90 ms RTT plus
encode/decode/buffering.

Translation: **great for single-player, RPGs, strategy, emulation, and remote
desktop; marginal for competitive FPS or fighting games.** No setting changes the
distance; just calibrate expectations. Lower resolution reduces encode/decode
time slightly but won't meaningfully move the network component.

## Ports / NAT

Tailscale (above) is the recommended path and needs no port config. It usually
negotiates a direct connection; worst case it relays through a DERP node and adds
a little latency.

If you'd rather forward ports on the server for the lowest possible latency, open
these to the gaming PC:

- **TCP:** 47984, 47989, 48010
- **UDP:** 47998, 47999, 48000, 48010

Keep the Sunshine web UI (TCP 47990) local — don't expose it. Note that
port-forwarding only helps if the *server* side has a reachable public IP; if the
Texas connection is itself behind CGNAT, Tailscale is mandatory, not optional.

## Quick start

1. **Texas PC:** install Sunshine, install Tailscale, sign in. In Sunshine's web
   UI (https://localhost:47990) confirm the NVENC/HEVC encoder is selected.
2. **Yukon client:** install Moonlight, install Tailscale, sign in to the same
   tailnet.
3. In Moonlight, add the host by its Tailscale IP; enter the PIN shown in
   Sunshine to pair.
4. In Moonlight Settings: codec **HEVC**, bitrate **~6 Mbit**, resolution
   **720p60** to start. Adjust from there.
5. Watch the 200 GB, not the 15 Mbit.
