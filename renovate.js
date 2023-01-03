module.exports = {
  globalExtends: ['github>jnewland/.github'],
  labels: ['renovate'],
  extends: ['config:base'],
  rebaseWhen: 'conflicted',
  binarySource: 'global',
};
