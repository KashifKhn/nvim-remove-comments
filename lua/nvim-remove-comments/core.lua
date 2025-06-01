local ts = vim.treesitter
local parsers = require("nvim-treesitter.parsers")

local M = {}

function M.remove_comments()
	local bufnr = vim.api.nvim_get_current_buf()
	local ft = vim.bo[bufnr].filetype

	if not parsers.has_parser(ft) then
		return
	end
	local parser = parsers.get_parser(bufnr, ft)

	if not parser then
		return
	end

	local root = parser:parse()[1]:root()
	local query = ts.query.parse(ft, [[ (comment) @comment ]])

	local regions = {}

	for id, node in query:iter_captures(root, bufnr, 0, -1) do
		local srow, scol, erow, ecol = node:range()
		table.insert(regions, 1, { srow, scol, erow, ecol })
	end

	for _, r in ipairs(regions) do
		vim.api.nvim_buf_set_text(bufnr, r[1], r[2], r[3], r[4], {})
	end
end

return M
