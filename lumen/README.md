# lumen

A calm one-touch browser game. You are a small light circling a glowing orb.
Let go, drift into the next light's ring, and make quiet music as the sky
travels from dusk through night and aurora to dawn.

## Play

Any static server works (no build step, no dependencies):

```sh
python3 -m http.server 8000 --directory lumen
# open http://localhost:8000
```

To play on your phone, run the server on your computer and open
`http://<your-computer's-LAN-IP>:8000` on the phone (same Wi-Fi). Use
"Add to Home Screen" for full-screen play.

## Controls

| Action | Touch | Desktop |
|---|---|---|
| Let go | tap anywhere | click, Space, Enter, W or Up |
| Pause | pause button | Esc or P |
| Sound | speaker button | M |

## Rules

- Each light you reach is +1. Fly straight at a light's heart for a **perfect** (+2).
- **Flow**: one light per journey; chase your best.
- **Zen**: no endings. A missed jump returns you to the last light.
- Stardust collects across all journeys and wakes new light colors.

## Files

- `index.html`, `style.css`: screens and HUD
- `game.js`: game loop (fixed 120 Hz step), world generation, rendering, Web Audio music
