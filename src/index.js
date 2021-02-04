import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import './style.css'
import './wasm_exec.js'

let fontSize = '24px';

const jsonEditor = monaco.editor.create(document.getElementById('jsonEditor'), {
	language: 'json',
	theme: 'vs-dark',
	automaticLayout: true,
	fontSize,
	minimap: {
		enabled: false
	},
	value: `// JSON data goes here\n{
"reasonsThisIsBad": ["no error handling", "limited tag support", "slow"]
}`
});

const hclEditor = monaco.editor.create(document.getElementById('hclEditor'), {
	language: 'hcl',
	theme: 'vs-dark',
	automaticLayout: true,
	minimap: {
		enabled: false
	},
	fontSize,
	value: ``
});

const output = document.getElementById('output');

let oldHcl = null;
let oldContext = null;

setInterval(function() {
	if (typeof parse_hcl === undefined) {
		return;
	}

	try {
		const hcl = hclEditor.getValue();
		const json = jsonEditor.getValue().replace(/\/\/.*/, '');
		if (oldHcl === hcl && oldContext === json) {
			return;
		}
		oldHcl = hcl;
		oldContext = json;
		const context = JSON.parse(json);

		console.log('Execution context', context);
		
		console.log('HCL', hcl);
		parse_hcl(hcl, context, output)
	} catch(e) {
		console.error(e);
	}
}, 1000);

