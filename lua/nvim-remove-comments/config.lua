local M = {}

M.queries = {
	javascript = [[ (comment) @comment ]],
	javascriptreact = [[ (comment) @comment ]],
	typescript = [[ (comment) @comment ]],
	typescriptreact = [[ (comment) @comment ]],
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
	dart = [[
  (comment) @comment
  (documentation_comment) @comment
]],
}

return M
