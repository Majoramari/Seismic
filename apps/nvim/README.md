# seismic.nvim

Automatic coding time tracking for [Neovim](https://neovim.io/), part of the [Seismic](https://seismic.icu) ecosystem.

Tracks your coding activity in the background and syncs it to your Seismic dashboard, same as the VS Code extension and
JetBrains plugin, so your stats stay unified across every editor you use.

## Features

- Automatic heartbeat tracking on edit, file switch, save, and focus
- Shares one API key across VS Code, JetBrains IDEs, and Neovim via `~/.seismic.cfg`
- Offline queueing — heartbeats that fail to send are retried automatically, up to 3 attempts
- Git branch detection
- Optional statusline integration (works with lualine or any statusline that accepts a Lua function component)

## Requirements

- Neovim 0.10+
- `curl` available on your `$PATH`
- `git` (optional, only needed for branch detection)

## Installation

This plugin lives inside the [Seismic monorepo](https://github.com/majoramari/seismic) at `apps/nvim`.

Lazy.nvim

```lua
{
  "majoramari/seismic",
  event = { "BufReadPost", "BufNewFile" },
  init = function(plugin)
    vim.opt.rtp:append(plugin.dir .. "/apps/nvim")
  end,
  main = "seismic",
  opts = {},
}
```

lazy.nvim will clone the monorepo, `init` points Neovim at the right subfolder, and `opts` triggers
`require("seismic").setup(opts)` automatically.

## Configuration

All options are optional — defaults shown below:

```lua
opts = {
  api_url = "https://correct-wolverine-majoramari-6049fd71.koyeb.app", -- production API
  enabled = true,          -- tracking on/off
  statusline_enabled = true, -- show status in statusline integrations
}
```

Your **API key** is never set here it's always read fresh from `~/.seismic.cfg`, the same shared config file used by
the VS Code extension and JetBrains plugin. Set it once with `:SeismicSetApiKey` in any editor and all three pick it up
automatically.

## Commands

| Command                 | Description                                               |
|-------------------------|-----------------------------------------------------------|
| `:SeismicSetApiKey`     | Prompts for your API key and saves it to `~/.seismic.cfg` |
| `:SeismicEnable`        | Resumes tracking                                          |
| `:SeismicDisable`       | Pauses tracking                                           |
| `:SeismicOpenDashboard` | Opens your Seismic dashboard in the browser               |
| `:SeismicStatus`        | Prints current tracking status                            |

## Statusline integration

`seismic.nvim` exposes a function you can plug into any statusline plugin. Example
for [lualine.nvim](https://github.com/nvim-lualine/lualine.nvim):

```lua
local function get_seismic()
  return require("seismic.statusline").get()
end

return {
  "nvim-lualine/lualine.nvim",
  opts = {
    sections = {
      lualine_x = {
        { get_seismic, icon = "󰥔" }, -- requires a Nerd Font; drop `icon` if you don't have one
        "encoding",
        "filetype",
      },
    },
  },
}
```

The status text updates once per minute and shows:

- `Set API Key` — no key configured yet
- `Paused` — tracking disabled via `:SeismicDisable`
- `Offline` — couldn't reach the API
- `2h 14m` / `< 1m` — today's tracked coding time

## How it works

Heartbeats fire on:

- Text changes (throttled to once per 2 minutes unless the file changed)
- Switching buffers
- Saving a file
- Regaining focus (subject to the same throttle)

Each heartbeat includes: file path, project (cwd basename), filetype, git branch, OS, hostname, cursor line, buffer line
count, and timezone — the same payload shape as the VS Code and JetBrains integrations, so stats aggregate consistently
across editors.

If a heartbeat fails to send (offline, API down), it's queued in memory and retried automatically every 5 minutes, up to
3 attempts before being dropped.

## License

GPL-3.0 — see [LICENSE](../../LICENSE) for details.