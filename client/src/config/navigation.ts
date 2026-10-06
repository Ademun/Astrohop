import type {Component} from 'vue'
import {Galaxy, Map, User} from '@lucide/vue'

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
    {label: 'My missions', icon: Map, to: '/missions'},
    {label: 'Object catalog', icon: Galaxy, to: '/catalog'},
    {label: 'Account', icon: User, to: '/account'}
]

export const footerNavItems: FooterItem[] = [
    {label: 'About', to: '/about'},
]
