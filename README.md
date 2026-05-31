<img src="https://bannermd.airopi.dev/banner?title=cup-notifier&desc=Notifications%20for%20cup&repo=AiroPi/cup-notifier" width="100%" alt="banner"/>

<!--p align="center">
  <a href="https://www.buymeacoffee.com/airopi" target="_blank">
    <img alt="Static Badge" src="https://img.shields.io/badge/Buy_me_a_coffee!-grey?style=for-the-badge&logo=buymeacoffee">
  </a>
</p-->

<h1 align="center">Cup notifier</h1>

A simple golang service to check whenever a new update is detected by [cup](https://github.com/sergi0g/cup). Can be executed as a Docker container.

## Deploy

Copy the compose file in [`deploy/`](./deploy/compose.yaml).

Add a `.env` file with the following variables:
- `CUP_URL`: the URL of your cup instance
- `NOTIFICATION_URLS`: a comma-separated list of [apprise notifications urls](https://github.com/caronc/apprise#supported-notifications)
- `INSECURE_SKIP_VERIFY`: set to `true` if your cup instance is behind an untrusted https endpoint

## Support, Feedback and Community

You can reach me over Discord at `@airo.pi`. Feel free to open an issue if you encounter any problem!

## How to contribute

I would ❤️ to see your contribution! Simply open a pull request 

## License

cup-notifier is under the MIT Licence.
