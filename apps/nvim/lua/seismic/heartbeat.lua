local M = {}

local config = require("seismic.config")
local detector = require("seismic.detector")
local queue = require("seismic.queue")

local HEARTBEAT_INTERVAL_SECONDS = 30

local last_heartbeat_time = 0
local last_file = ""
local has_shown_invalid_key_warning = false
local keystroke_count = 0

function M.record_keystroke()
	keystroke_count = keystroke_count + 1
end

local function build_payload(bufnr)
	local lines = vim.api.nvim_buf_line_count(bufnr)
	local cursor_line = vim.api.nvim_win_get_cursor(0)[1]

	return {
		file = vim.api.nvim_buf_get_name(bufnr),
		project = detector.detect_project(config.use_git_root_project_name()),
		language = detector.language_id(bufnr),
		editor = "neovim",
		branch = detector.detect_branch(),
		os = detector.detect_os(),
		machine = detector.detect_machine(),
		lines = lines,
		keystrokes = keystroke_count,
		cursorLine = cursor_line,
		timezone = detector.detect_timezone(),
		time = os.time() * 1000,
	}
end

local function notify_invalid_key()
	if has_shown_invalid_key_warning then
		return
	end
	has_shown_invalid_key_warning = true
	vim.notify("Seismic: Invalid API key. Run :SeismicSetApiKey to update it.", vim.log.levels.WARN)
end

local function send(payload)
	local api_key = config.get_api_key()
	local api_url = config.get_api_url()
	local body = vim.json.encode(payload)

	vim.system({
		"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
		"-X", "POST",
		"-H", "Content-Type: application/json",
		"-H", "Authorization: Bearer " .. api_key,
		"-d", body,
		api_url .. "/api/heartbeat",
	}, { text = true }, function(result)
		local code = result.stdout

		if code == "401" then
			notify_invalid_key()
			return
		end

		if code:match("^2%d%d") then
			queue.flush(api_key, api_url)
		else
			queue.enqueue(payload)
		end
	end)
end

function M.handle_activity(bufnr, forced)
	if not config.is_enabled() then
		return
	end
	if not config.has_api_key() then
		return
	end
	if not detector.should_track(bufnr) then
		return
	end

	local now = os.time()
	local filename = vim.api.nvim_buf_get_name(bufnr)
	local file_changed = filename ~= last_file

	if not forced and not file_changed and (now - last_heartbeat_time) < HEARTBEAT_INTERVAL_SECONDS then
		return
	end

	last_heartbeat_time = now
	last_file = filename

	local payload = build_payload(bufnr)
	keystroke_count = 0
	send(payload)
end

function M.flush_queue()
	queue.flush(config.get_api_key(), config.get_api_url())
end

return M
