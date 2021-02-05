import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import './style.css'
import './wasm_exec.js'

let fontSize = '24px';

let defaultJSON = `// This JSON data will be available to 
// HCL
{
"reasonsThisIsBad": ["lacks error handling", "has limited tag support", "is slow"]
}`;

let defaultHCL = `h1 {
    marquee {
        innerText = "This is a bad idea"
    }
}

p {
    innerText = "Among other things, it..."
}

ul {
    dynamic "li" {
        for_each = reasonsThisIsBad
        innerText = "...\${for_each}"
        style = "color: red"
    }
}
p {
    innerText = "You should only use it if you have at least \${length(reasonsThisIsBad) * 123} good reasons."
}`;

try {
	let loadedDoc = JSON.parse(decodeURI(window.location.hash.substr(1)));	
	defaultJSON = loadedDoc.json;
	defaultHCL = loadedDoc.hcl;
} catch(e) {
	console.warn("Error loading hcl/json from hash", e);
}


const jsonEditor = monaco.editor.create(document.getElementById('jsonEditor'), {
	language: 'json',
	theme: 'vs-dark',
	automaticLayout: true,
	fontSize,
	minimap: {
		enabled: false
	},
	value: defaultJSON
});

const hclEditor = monaco.editor.create(document.getElementById('hclEditor'), {
	language: 'hcl',
	theme: 'vs-dark',
	automaticLayout: true,
	minimap: {
		enabled: false
	},
	fontSize,
	value: defaultHCL
});

const output = document.getElementById('output');

let oldHcl = null;
let oldContext = null;

setInterval(function() {
	// It takes a while for the go wasm to download, parse, and execute. Until that happens, any functions it exposes won't be defined
	if (typeof parse_hcl === undefined) {
		return;
	}

	try {
		const hcl = hclEditor.getValue();
		let json = jsonEditor.getValue();
		if (oldHcl === hcl && oldContext === json) {
			return;
		}
		oldHcl = hcl;
		oldContext = json;

		window.location.hash = JSON.stringify({hcl, json});

		// JSON doesn't support comments, but it's nice to put a comment explaining the top window. This
		// strips out any comments to avoid angering JSON.stringify
		json = json.replaceAll(/\/\/.*/g, '')
		
		const context = JSON.parse(json);

		console.log('Execution context', context);
		
		console.log('HCL', hcl);
		parse_hcl(hcl, context, output)
	} catch(e) {
		console.error(e);
	}
}, 1000);

