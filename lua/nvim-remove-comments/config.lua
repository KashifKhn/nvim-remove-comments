local M = {}

M.queries = {
	javascript = [[ (comment) @comment ]],
	typescript = [[ (comment) @comment ]],
	java = [[
    (line_comment) @comment
    (block_comment) @comment
  ]],
	lua = [[ (comment) @comment ]],
	python = [[ (comment) @comment ]],
	go = [[ (comment) @comment ]],
	c = [[
    (line_comment) @comment
    (block_comment) @comment
  ]],
	cpp = [[
    (line_comment) @comment
    (block_comment) @comment
  ]],
	rust = [[ (line_comment) @comment ]],
	html = [[ (comment) @comment ]],
	css = [[ (comment) @comment ]],
	yaml = [[ (comment) @comment ]],
	toml = [[ (comment) @comment ]],
	bash = [[ (comment) @comment ]],
	sh = [[ (comment) @comment ]],
}

return M
