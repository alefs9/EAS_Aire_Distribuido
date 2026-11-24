import {datasetSelection} from './selection.js';
// Read-only UI handoff. This is not a POST /jobs body and never starts a job.
export function getSelectedDataset(){return datasetSelection.current();}
export function executionChoice(selection){
 const dataset=selection.current();
 return dataset ? Object.freeze({datasetId:dataset.id,period:dataset.period}) : null;
}
export function renderSelectionActions(shell,selection,onClear){
 const choice=executionChoice(selection);
 const control=document.createElement(choice?'a':'button');
 control.className='button button-primary';control.textContent='Continuar a ejecuciones';
 if(choice){control.href='/ejecuciones';control.dataset.route='';}
 else{control.type='button';control.disabled=true;}
 shell.actions.replaceChildren(control);
 if(choice){const clear=document.createElement('button');clear.type='button';clear.className='button button-secondary';clear.textContent='Quitar selección';clear.addEventListener('click',onClear);shell.actions.append(clear);}
}
