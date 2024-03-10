# 🎖️ Introduction

The purpose of this assignment is to evaluate the structure of the delivered application, the best practices you follow, along with your coding principles and skills.

# 🖥️ My Wonderful API

As part of this assignment you will create two requests. One that will seed your backend with 5,000 entries from [Randomuser.com API](https://randomuser.me/), and one that will fetch the results based on a few parameters.

## 📑 Requirements

- Create a `POST /populate` request, that adds 5000 random user entries from the [Randomuser.com API](https://randomuser.me/) to your server. For each user, store their name, email, phone number, picture, and registration date.
- Create a `GET /wonderfuls` request, that returns a list of your users. Return **10 users** sorted by registration date, with the most recent users appearing first.
  - Allow the following optional parameters:
    - **limit**
      - A limit on the number of users to be returned. `limit` should range between 1 and 100.
    - **starting_after**
      - A cursor for use in pagination. `starting_after` is a user ID that defines your place in the list. For instance, if you make a list request and receive 100 objects, ending with `ID=obj_foo`, your subsequent call can include `starting_after=obj_foo` in order to fetch the next page of the list.
    - **ending_before**
      - A cursor for use in pagination. `ending_before` is a user ID that defines your place in the list. For instance, if you make a list request and receive 100 objects, starting with `ID=obj_bar`, your subsequent call can include `ending_before=obj_bar` in order to fetch the previous page of the list.
    - **email**
      - A case-insensitive filter on the list based on the user's **`email`** field.
- Include some basic testing.
- Version your code with Git.
- Use [Docker Compose](https://docs.docker.com/compose/) to setup your backend, so that we can easily reproduce your project and test it.
- While you can choose any language to implement this assignment, we encourage you to consider using Go. This will allow us to better evaluate your compatibility with our existing codebase and development practices.

## 🔍 **The Review**

Please include a **README** file with the necessary instructions and the points you'd like to highlight. It will help us during the review, where we'll look at:

- The project structure and architecture
- Good coding practices
- Performance
- Code styling and formatting
- Naming and conventions
- VC history

**🎉 Have fun! 🎉**
