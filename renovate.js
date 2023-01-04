module.exports = {
  extends: ['github>urcomputeringpal/.github'],
  hostRules: [
    {
      hostType: 'github',
      matchHost: 'github.com',
      username: process.env.RENOVATE_USERNAME,
      token: process.env.RENOVATE_GITHUB_COM_TOKEN,
    },
  ],
  labels: ['renovate'],
  extends: ['config:base'],
  rebaseWhen: 'conflicted',
  binarySource: 'global',
};
