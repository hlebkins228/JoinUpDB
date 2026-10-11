.DEFAULT_GOAL := report

.PHONY: title report presentation

title:
	latexmk -pdf -interaction=nonstopmode -halt-on-error title.tex

report: title
	latexmk -pdf -interaction=nonstopmode -halt-on-error report.tex

presentation:
	latexmk -pdf -interaction=nonstopmode -halt-on-error presentation.tex
