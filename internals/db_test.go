package internals

import "testing"

func Test_Functions(tf *testing.T) {

	var conn DBConn

	// open a tcp connection to the database
	conn.Setup()

	tf.Run("getBlogsByAuthorID", func(t *testing.T) {

		blogs, err := conn.getBlogsByAuthor(2)
		if err != nil {
			t.Error(err)
		}

		for _, blog := range blogs {
			if blog.AuthorID != 2 {
				t.Error("blog is not by written by authorID=2")
			}
		}

	})

}
