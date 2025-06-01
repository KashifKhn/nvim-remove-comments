local core = require("nvim-remove-comments.core")

local M = {}

function M.setup(opts)
	vim.keymap.set("n", "<leader>rc", function()
		M.remove_comments()
	end, { desc = "Remove Comments" })
end

function M.remove_comments()
	core.remove_comments()
end

return M
