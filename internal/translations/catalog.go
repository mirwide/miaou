// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package translations

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"en_US": &dictionary{index: en_USIndex, data: en_USData},
		"ru_RU": &dictionary{index: ru_RUIndex, data: ru_RUData},
	}
	fallback := language.MustParse("ru-RU")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"Возникла ошибка, попробуй позже.":                  0,
	"Изображения не поддерживаюются в этой версии LLM.": 3,
	"Привет! Расскажи кратко что ты умеешь.":            4,
	"Слишком много запросов, попробуй позже.":           2,
	"Текущая модель %s.":                                1,
	"Текущая модель %s. Сменить:":                       5,
}

var en_USIndex = []uint32{ // 7 elements
	0x00000000, 0x0000002b, 0x00000040, 0x0000006a,
	0x0000009b, 0x000000c0, 0x000000dc,
} // Size: 52 bytes

const en_USData string = "" + // Size: 220 bytes
	"\x02An error occurred, please try again later.\x02Current model %[1]s." +
	"\x02To many requests, please try again later.\x02Images are not supporte" +
	"d in this version of LLM.\x02Hi! Briefly tell me what you can do.\x02Cur" +
	"ren model %[1]s. Change:"

var ru_RUIndex = []uint32{ // 7 elements
	0x00000000, 0x0000003c, 0x0000005f, 0x000000a8,
	0x00000101, 0x00000147, 0x0000017a,
} // Size: 52 bytes

const ru_RUData string = "" + // Size: 378 bytes
	"\x02Возникла ошибка, попробуй позже.\x02Текущая модель %[1]s.\x02Слишком" +
	" много запросов, попробуй позже.\x02Изображения не поддерживаюются в это" +
	"й версии LLM.\x02Привет! Расскажи кратко что ты умеешь.\x02Текущая моде" +
	"ль %[1]s. Сменить:"

	// Total table size 702 bytes (0KiB); checksum: 4CC82EA0
