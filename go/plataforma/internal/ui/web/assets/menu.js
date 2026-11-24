// Accessible disclosure navigation; it is not an ARIA application menu.
export function bindMenu(toggle, navigation, root = document, viewport = window.matchMedia('(max-width: 700px)')) {
 const isOpen = () => toggle.getAttribute('aria-expanded') === 'true';
 const setOpen = open => {toggle.setAttribute('aria-expanded', String(open)); navigation.classList.toggle('is-open', open);};
 const close = (restoreFocus = false) => {setOpen(false); if (restoreFocus) toggle.focus();};
 toggle.addEventListener('click', () => setOpen(!isOpen()));
 toggle.addEventListener('keydown', event => {
  if (event.key === 'ArrowDown') {event.preventDefault(); setOpen(true); navigation.querySelector('a')?.focus();}
 });
 root.addEventListener('keydown', event => {
  if (event.key === 'Escape' && isOpen()) {event.preventDefault(); close(navigation.contains(root.activeElement) || root.activeElement === toggle);}
 });
 root.addEventListener('click', event => {
  if (isOpen() && !navigation.contains(event.target) && !toggle.contains(event.target)) close();
 });
 viewport.addEventListener('change', () => close());
 return close;
}
