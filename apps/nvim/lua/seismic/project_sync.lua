local M = {}

local config = require("seismic.config")
local detector = require("seismic.detector")

local last_sync = 0
local MIN_SYNC_INTERVAL_SECONDS = 15

local function send(payload)
	local api_key = config.get_api_key()
	local api_url = config.get_api_url()
	local body = vim.json.encode(payload)

	vim.system({
		"curl", "-s", "-o", "/dev/null",
		"-X", "POST",
		"-H", "Content-Type: application/json",
		"-H", "Authorization: Bearer " .. api_key,
		"-d", body,
		api_url .. "/api/projects/sync",
	}, { text = true })
end

function M.sync(force)
	if not config.is_enabled() or not config.has_api_key() then
		return
	end

	local now = os.time()
	if not force and now - last_sync < MIN_SYNC_INTERVAL_SECONDS then
		return
	end

	config.refresh_editor_settings()
	local metadata = detector.detect_project_metadata(config.use_git_root_project_name())
	if not metadata.repoUrl and #metadata.commits == 0 and not metadata.websiteUrl then
		return
	end

	last_sync = now
	send(metadata)
end

return M
