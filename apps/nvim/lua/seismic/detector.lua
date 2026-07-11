local M = {}

function M.detect_project()
	local cwd = vim.fn.getcwd()
	return vim.fn.fnamemodify(cwd, ":t")
end

function M.detect_branch()
	local result = vim.fn.systemlist("git rev-parse --abbrev-ref HEAD 2>/dev/null")
	if vim.v.shell_error ~= 0 or #result == 0 then
		return nil
	end
	return result[1]
end

function M.detect_os()
	local sysname = vim.uv.os_uname().sysname
	if sysname == "Linux" then
		return "linux"
	end
	if sysname == "Darwin" then
		return "darwin"
	end
	if sysname:match("Windows") then
		return "windows"
	end
	return sysname:lower()
end

function M.detect_machine()
	return vim.uv.os_gethostname()
end

function M.detect_timezone()
	local tz = os.date("%Z")
	return tz or ""
end

function M.language_id(bufnr)
	local ft = vim.bo[bufnr].filetype
	return ft ~= "" and ft or "text"
end

function M.should_track(bufnr)
	if vim.bo[bufnr].buftype ~= "" then
		return false
	end -- skip terminal, quickfix, etc
	local name = vim.api.nvim_buf_get_name(bufnr)
	return name ~= ""
end

return M