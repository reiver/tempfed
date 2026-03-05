package librefs

import (
	gourl "net/url"
)

func RelayActor(host string) string {
	var url = gourl.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/actor",
	}

	return url.String()
}

func RelayActorInBox(host string) string {
	var url = gourl.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/inbox",
	}

	return url.String()
}

func RelayActorFollowers(host string) string {
	return RelayActor(host) + "/followers"
}

func RelayActorFollowing(host string) string {
	return RelayActor(host) + "/following"
}
