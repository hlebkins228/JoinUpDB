// Code generated from api/openapi.yaml by go generate. DO NOT EDIT.

package controller

import "github.com/labstack/echo/v5"

const v2Prefix = "/api/v2"

type v2Operations struct {
	listCategories      echo.HandlerFunc
	listEvents          echo.HandlerFunc
	createEvent         echo.HandlerFunc
	getEvent            echo.HandlerFunc
	replaceEvent        echo.HandlerFunc
	patchEvent          echo.HandlerFunc
	deleteEvent         echo.HandlerFunc
	putEventCategory    echo.HandlerFunc
	createEventImage    echo.HandlerFunc
	putParticipation    echo.HandlerFunc
	deleteParticipation echo.HandlerFunc
	listReviews         echo.HandlerFunc
	createReview        echo.HandlerFunc
	createImage         echo.HandlerFunc
	getImage            echo.HandlerFunc
	createToken         echo.HandlerFunc
	createUser          echo.HandlerFunc
	replaceCurrentUser  echo.HandlerFunc
	listCreatedEvents   echo.HandlerFunc
	listJoinedEvents    echo.HandlerFunc
	getUser             echo.HandlerFunc
}

func registerV2(e *echo.Echo, handlers v2Operations, auth echo.MiddlewareFunc, requests *v2Requests) {
	group := e.Group(v2Prefix)
	group.Add("GET", "/categories", handlers.listCategories)
	group.Add("GET", "/events", handlers.listEvents)
	group.Add("POST", "/events", handlers.createEvent, auth, requests.validate("application/json", "#/components/schemas/EventCreate", true, 422, 400))
	group.Add("GET", "/events/:id", handlers.getEvent)
	group.Add("PUT", "/events/:id", handlers.replaceEvent, auth, requests.validate("application/json", "#/components/schemas/EventReplace", true, 422, 400))
	group.Add("PATCH", "/events/:id", handlers.patchEvent, auth, requests.validate("application/merge-patch+json", "#/components/schemas/EventPatch", true, 422, 415))
	group.Add("DELETE", "/events/:id", handlers.deleteEvent, auth)
	group.Add("PUT", "/events/:id/categories/:categoryId", handlers.putEventCategory, auth)
	group.Add("POST", "/events/:id/images", handlers.createEventImage, auth, requests.validate("multipart/form-data", "#/components/schemas/ImageUpload", true, 422, 400))
	group.Add("PUT", "/events/:id/participants/me", handlers.putParticipation, auth)
	group.Add("DELETE", "/events/:id/participants/me", handlers.deleteParticipation, auth)
	group.Add("GET", "/events/:id/reviews", handlers.listReviews)
	group.Add("POST", "/events/:id/reviews", handlers.createReview, auth, requests.validate("application/json", "#/components/schemas/ReviewCreate", true, 400, 400))
	group.Add("POST", "/images", handlers.createImage, auth, requests.validate("multipart/form-data", "#/components/schemas/ImageUpload", true, 400, 400))
	group.Add("GET", "/images/:id", handlers.getImage)
	group.Add("POST", "/tokens", handlers.createToken, requests.validate("application/json", "#/components/schemas/Credentials", false, 422, 400, "login", "password"))
	group.Add("POST", "/users", handlers.createUser, requests.validate("application/json", "#/components/schemas/UserCreate", true, 422, 400))
	group.Add("PUT", "/users/me", handlers.replaceCurrentUser, auth, requests.validate("application/json", "#/components/schemas/UserReplace", true, 422, 400))
	group.Add("GET", "/users/me/events", handlers.listCreatedEvents, auth)
	group.Add("GET", "/users/me/participations", handlers.listJoinedEvents, auth)
	group.Add("GET", "/users/:id", handlers.getUser, auth)
}
