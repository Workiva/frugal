

const System SYSTEM_FOO = "foo"
const System SYSTEM_BAR = "bar"
const System SYSTEM_BAZ = "baz"

const list<map<i64,string>> COMPLEX_LIST = [
	{
		1: "foo",
	},
]

const list<list<i64>> LIST_OF_LISTS = [
	[
		1,
		2,
		3,
	],
	[
		4,
		5,
		6,
	],
]

/**
 * this is a docstring
 */
const map<string,list<i32>> MAP_OF_LISTS = {
	"foo": [
		1,
		2,
		3,
	],
	"bar": [
		5,
		6,
		7,
	],
	"baz": [
		8,
		9,
		10,
	],
}

const map<System,Schema> REGISTERED_SCHEMAS = {
	SYSTEM_FOO: {
		"name": SYSTEM_FOO,
		"properties": [
			{
				"id": "prop1",
				"title": "property 1",
				"inScope": true,
			},
			{
				"id": "prop2",
				"title": "property 2",
			},
		],
	},
	"bar": {
		"name": SYSTEM_BAR,
		"properties": [
			{
				"id": "prop3",
				"title": "property 3",
				"inScope": true,
			},
			{
				"id": "prop4",
				"title": "property 4",
			},
		],
	},
}

const list<i64> SIMPLE_LIST_INT = [
	0,
	1,
	2,
	3,
]

const map<i64,string> SIMPLE_MAP_INT_KEYS = {
	0: "foo",
	1: "bar",
	2: "baz",
}

const map<string,i64> SIMPLE_MAP_STRING_KEYS = {
	"foo": 0,
	"bar": 1,
	"baz": 2,
}

const set<string> SIMPLE_SET_STRING = [
	"foo",
	"bar",
	"baz",
]

typedef string System

struct SchemaProperty {
	1: string id,
	2: string title,
	3: bool inScope,
}

struct Schema {
	1: required string foo,
	2: optional list<SchemaProperty> properties,
}

