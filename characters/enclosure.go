// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

func knownEnclosures() []Character {
	return []Character{
		Character_QUOTATION_MARK_TRIPLE,
		Character_QUOTATION_MARK_DOUBLE,
		Character_QUOTATION_MARK,
		Character_APOSTROPHE,
		Character_ASTERISK,

		// Character_EMPTY has to be last as it can ALWAYS appear
		Character_EMPTY,
	}
}

func GetEnclosureNames() []string {
	return getNames(knownEnclosures())
}

func DetectEnclosure(data []byte) (Character, int) {
	return detect(data, knownEnclosures())
}
