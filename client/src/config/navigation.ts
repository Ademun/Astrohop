import type { Component } from 'vue'
import { Map, Galaxy } from '@lucide/vue'

export interface NavItem {
  label: string
  icon: Component
  to?: string
}

export interface FooterItem {
  label: string
  to: string
}

export const primaryNavItems: NavItem[] = [
  { label: 'My missions', icon: Map, to: '/missions' },
  { label: 'Object registry', icon: Galaxy, to: '/objects' },
]

export const footerNavItems: FooterItem[] = [
  { label: 'FAQ', to: '/faq' },
  { label: 'About', to: '/about' },
  { label: 'Donate', to: '/donate' },
]
