## AI Blog

This web application will be a blog I'll use to publish and update my followers on my AI studies.

- It will have an admin area that only I will have access to, but in the future I plan to add the option to grant access to other people to publish articles.
- In the admin area, only those who can log in — initially just me — will be able to create posts using a text editor in a dedicated post creation tab.
- I want a dashboard with access metrics and post reading stats, such as: most-read post, how many accesses each post received.
- I want the blog design to be pleasant, with colors and fonts that make the reader feel comfortable while reading, without losing the tech essence.

## Deployment

I'll be installing this on a Hostinger virtual machine, but before pushing to production I want to test the blog locally, so it needs a setup to run locally. Since I'll initially deploy it on a virtual machine using Docker, I need a lean, production-ready Docker configuration.

## Tech Stack

- The backend will use the latest version of Golang and will be built as an API following proper Golang conventions, RESTful routes/resources, no external database — using SQLite3 if needed.
- The frontend can be built using Golang with Fiber.
- When opting for external libraries, use those considered most recommended and secure. (Only opt for an external lib if it's truly the only or best solution.)
- Security restrictions should be less strict during development so I can run tests (make this configurable and document it in the readme file).
- It would be interesting to add security-focused tests, such as verifying that rate limits on important endpoints, bandwidth, and TTLs are working correctly (these could be integration tests).
- Always update the README file with important configuration aspects.
- Always check for proper HTTP headers, such as CSP, etc.

## Use Cases

The application must store all access/security-related data in encrypted form.

To access the admin panel, it could be via URL by typing something like `/tech`, or if you have a better suggestion. I want to find a route that isn't too obvious.

I want to be able to categorize posts by category, and it would be interesting to have a timeline page of my posts.

The homepage needs to be aesthetic and organized.

Take care of exposed endpoints — configure them to block bot behavior or brute force, and apply rate limiting at the necessary entry points.

The editor could be Markdown-based and, if possible, dynamic — allowing the addition of images, videos, and maybe audio.

The admin internal area could be divided following the standard sidebar pattern — initially something like: all posts, metrics, and within each of those tabs, the implied sub-functions.

## New Ideas

Elaborate a plan that takes all these requirements into consideration and suggest important features you consider relevant for a service like this.