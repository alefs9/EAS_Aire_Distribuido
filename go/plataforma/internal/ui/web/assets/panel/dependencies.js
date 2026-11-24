export class PanelDependencyUnavailable extends Error {
 constructor(resource){super('La fuente de información no está conectada.');this.name='PanelDependencyUnavailable';this.resource=resource;}
}
