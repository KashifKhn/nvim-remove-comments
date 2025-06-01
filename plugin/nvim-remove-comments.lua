if vim.g.loaded_nvim_remove_comments then
	return
end
vim.g.loaded_nvim_remove_comments = true

require("nvim-remove-comments").setup()
