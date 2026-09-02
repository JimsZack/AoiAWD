import api from '../index'

export const ping = () => api.get('/ping')

export const getInfo = () => api.get('/info')

export const getPlugins = () => api.get('/listplugin')

export const reloadPlugins = () => api.get('/reloadplugin')
