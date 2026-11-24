// Plain text only: callers must never pass backend text to innerHTML.
export function messageState(kind, title, description, retry) {
  const section = document.createElement('section');
  section.className = `message message-${kind}`;
  section.setAttribute('role', kind === 'error' ? 'alert' : 'status');
  if (kind === 'loading') section.setAttribute('aria-busy', 'true');
  const heading = document.createElement('h2'); heading.textContent = title;
  const text = document.createElement('p'); text.textContent = description;
  section.append(heading, text);
  if (retry) {const button = document.createElement('button'); button.type = 'button'; button.className = 'button button-secondary'; button.textContent = 'Volver a intentar'; button.addEventListener('click', retry); section.append(button);}
  return section;
}
export const loadingState = (description = 'Espera un momento, por favor.') => messageState('loading', 'Cargando información', description);
export const emptyState = (title = 'Sin datos disponibles', description = 'Todavía no hay información para mostrar.') => messageState('empty', title, description);
export const errorState = (description = 'No se pudo obtener la información. Intenta nuevamente.', retry) => messageState('error', 'No pudimos completar la consulta', description, retry);
