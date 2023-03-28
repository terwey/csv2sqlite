// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package characters

func KnownNewlines() []Character {
	return []Character{
		Character_CARRIAGE_RETURN_LINE_FEED,
		Character_LINE_FEED,
		Character_CARRIAGE_RETURN,
	}
}

func GetNewlineNames() []string {
	return getNames(KnownNewlines())
}

func DetectNewline(data []byte) Character {
	nl, _ := detect(data, KnownNewlines())
	return nl
}
