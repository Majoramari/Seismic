local M = {}

local function systemlist(args)
	local result = vim.fn.systemlist(args)
	if vim.v.shell_error ~= 0 or #result == 0 then
		return nil
	end
	return result
end

function M.detect_git_root()
	local result = systemlist({ "git", "rev-parse", "--show-toplevel" })
	return result and result[1] or nil
end

function M.detect_project(use_git_root)
	if use_git_root then
		local git_root = M.detect_git_root()
		if git_root then
			return vim.fn.fnamemodify(git_root, ":t")
		end
	end

	local cwd = vim.fn.getcwd()
	return vim.fn.fnamemodify(cwd, ":t")
end

function M.detect_branch()
	local result = systemlist({ "git", "rev-parse", "--abbrev-ref", "HEAD" })
	return result and result[1] or nil
end

function M.detect_repo_url(project_path)
	local result = systemlist({ "git", "-C", project_path, "config", "--get", "remote.origin.url" })
	local url = result and result[1] or nil
	if not url or url == "" then
		return nil
	end
	local host, repo = url:match("^git@([^:]+):(.+)$")
	if host and repo then
		return "https://" .. host .. "/" .. repo:gsub("%.git$", "")
	end
	return url:gsub("%.git$", "")
end

function M.detect_website_url(project_path)
	local package_json = project_path .. "/package.json"
	if vim.fn.filereadable(package_json) ~= 1 then
		return nil
	end

	local ok, body = pcall(vim.json.decode, table.concat(vim.fn.readfile(package_json), "\n"))
	if not ok or type(body) ~= "table" then
		return nil
	end
	return body.homepage or body.website
end

function M.detect_commits(project_path, limit)
	limit = limit or 20
	local result = systemlist({
		"git",
		"-C",
		project_path,
		"log",
		"-" .. tostring(limit),
		"--format=%H%x1f%ct%x1f%an%x1f%ae%x1f%s",
	})
	if not result then
		return {}
	end

	local commits = {}
	for _, line in ipairs(result) do
		local parts = vim.split(line, "\31", { plain = true })
		table.insert(commits, {
			hash = parts[1],
			committedAt = parts[2] and tonumber(parts[2]) and tonumber(parts[2]) * 1000 or nil,
			authorName = parts[3],
			authorEmail = parts[4],
			message = parts[5],
		})
	end
	return commits
end

function M.detect_project_metadata(use_git_root)
	local cwd = vim.fn.getcwd()
	local project_path = use_git_root and M.detect_git_root() or nil
	project_path = project_path or cwd
	local commits = M.detect_commits(project_path)

	return {
		project = vim.fn.fnamemodify(project_path, ":t"),
		repoUrl = M.detect_repo_url(project_path),
		websiteUrl = M.detect_website_url(project_path),
		lastCommitAt = commits[1] and commits[1].committedAt or nil,
		commits = commits,
	}
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
