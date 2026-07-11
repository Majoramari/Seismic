local M = {}

local config = require("seismic.config")

local cached_text = ""
local last_fetch = 0
local FETCH_INTERVAL = 60

local function format_seconds(seconds)
	if seconds < 60 then
		return "< 1m"
	end
	local hours = math.floor(seconds / 3600)
	local minutes = math.floor((seconds % 3600) / 60)
	if hours > 0 then
		return string.format("%dh %dm", hours, minutes)
	end
	return string.format("%dm", minutes)
end

local function refresh()
	if not config.options.statusline_enabled then
		cached_text = ""
		return
	end
	if not config.has_api_key() then
		cached_text = "Set API Key"
		return
	end
	if not config.is_enabled() then
		cached_text = "Paused"
		return
	end

	local api_key = config.get_api_key()
	local api_url = config.get_api_url()

	vim.system({
		"curl", "-s",
		"-H", "Authorization: Bearer " .. api_key,
		api_url .. "/api/stats/summary?range=today",
	}, { text = true }, function(result)
		if result.code ~= 0 or not result.stdout or result.stdout == "" then
			cached_text = "Offline"
			return
		end
		local ok, decoded = pcall(vim.json.decode, result.stdout)
		if not ok or not decoded.data then
			cached_text = "Offline"
			return
		end
		cached_text = format_seconds(decoded.data.totalSeconds)
	end)
end

function M.get()
	local now = os.time()
	if now - last_fetch > FETCH_INTERVAL then
		last_fetch = now
		refresh()
	end
	return cached_text
end

return M
