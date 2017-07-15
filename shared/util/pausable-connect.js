import {createConnect} from 'react-redux/src/connect/connect'
import defaultSelectorFactory from 'react-redux/src/connect/selectorFactory'

function selectorFactory(dispatch, factoryOptions) {
  const selector = defaultSelectorFactory(dispatch, factoryOptions)
  let cachedResult
  const pausableSelector = function(state, ownProps) {
    if (ownProps.isActive) {
      cachedResult = selector(state, ownProps)
    }
    return cachedResult
  }
  return pausableSelector
}

export default createConnect({selectorFactory})
