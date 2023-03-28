// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Yorick Terweijden

package knownconfigs

import "github.com/terwey/csv2sqlite/characters"

// CSV returns a config set for COMMA, QUOTATION_MARK and LF
func CSV() characters.Config {
	c := characters.NewConfig(characters.Character_COMMA, characters.Character_QUOTATION_MARK)
	c.EOL = characters.Character_LINE_FEED

	return c
}

// TSV returns a config set for TAB, QUOTATION_MARK and LF
func TSV() characters.Config {
	c := characters.NewConfig(characters.Character_TAB, characters.Character_QUOTATION_MARK)
	c.EOL = characters.Character_LINE_FEED

	return c
}

// PipeSV returns a config set PIPE TAB, QUOTATION_MARK and LF
func PipeSV() characters.Config {
	c := characters.NewConfig(characters.Character_PIPE, characters.Character_QUOTATION_MARK)
	c.EOL = characters.Character_LINE_FEED

	return c
}
