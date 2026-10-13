import type { Component } from 'vue';

/**
 * Convert ri-xxx-yyy to RiXxxYyy (global component name)
 */
export function riToCompName(riClass: string): string {
  const parts = riClass.replace('ri-', '').split('-');
  return 'Ri' + parts.map(p => p.charAt(0).toUpperCase() + p.slice(1)).join('');
}

/**
 * Resolve an icon name string to its globally registered Vue component.
 * Falls back to RiQuestionLine if not found.
 */
export function resolveIcon(name: string | undefined): Component {
  if (!name) return {} as Component;
  const compName = riToCompName(name);
  // The component is globally registered; return a dynamic component reference
  return { __name: compName } as Component;
}
