# CourseSelectionSystem
A toy project that serves as a middle school course selection system.

### Behavior
The project view all users in three levels: the monitor, teachers and students.
The operations and the priviledge requiring at least are as follows:
1. Account System:
   1. Register[Monitor]: create a new account and set its name, identity and initial password. The name and password should be [a-zA-Z0-9_]* in 10 words,
   and temporarily the identity includes the user's previledge and class it is in.
   2. Remove[Monitor]: erase a account from the whole system, actually behaving
   as enforcing the user to log out immediately and have no ability to come back.
   3. LogIn[Student]: use account and password to log into the system.
   4. LogOut[Student]: log out from the system.
   5. ModifyPassword[Student]: anyone in the system can modify its own password.
   6. GetAllUsersInfo[Monitor]: list every user with name, password and identical information.
   7. GetPartUsersInfo[Teacher]: list part of users with a keyword of either class or course they are in.  
2. Course Selection System:  
   1. AddCourse[Monitor]: add a new course with initial info, including name,professor and maximum students.
   2. ModifyCourse[Monitor]: modify information of a course.
   3. LaunchCourse[Monitor]: make a elective course avaliable to students, both seats and information.
   4. SelectCourse[Student]: choose a course whose places are enough when no picked one.
   5. DropCourse[Student]: abandon a selected course.
3. Logging System: Only the monitor can view the behavior of every one.

### Designing
We seperate the whole project into frontend and backend. Because of Go's wonderful network framework, I choose to handle HTTP requests by net/http
package and use go routine implement the concurrency logic. Now the core logic
can be sealing into a handle() function that input a http request(more specifically Go's http.request struct) and output reply in the same format.
While the frontend is not my deal.

### API Protocol
We design the Web communication in a hidden-backend way, which means that the frontend only is able to know the single URL and a protocol regulating json in HTTP for communication, instead of using URL suffix for routing.  
1. All kinds of request are consists of three main parts, sealing in a HTTP request and with a POST method:   
   1. Identity information: all requests except LogIn should carry a unique authentation token as ID.
   2. Commands: the kind of request and corresponding parameters:
      1. Register:
      {
         "name":
         "password":
         "identityInfo":{
            "class":{"grade":,"class":}
            "previledge":
         }
      }
      2. Remove:
      {
         "name":
      }
      3. LogIn:   
      {
         "name":
         "password":
      }
      4. LogOut:   
      {
         null
      }
      5. ModifyPassword:   
      {
         "password":
      }
      6. GetAllUsersInfo:   
      {
         null
      }
      7. GetPartUsersInfo:   
      {
         "way": 0 or 1, representing by class or course.
         "class":{"grade":,"class":}
         "courseName":
      }
      8. AddCourse:   
      {
         "courseInfo":{
            "name":
            "teacherName":
            "maximum":
         }
      }
      9. ModifyCourse:   
      {
         "courseName":
         "courseInfo":{
            "name":
            "teacherName":
            "maximum":
         }
      }
      10. LaunchCourse:   
      {
         "courseName":
      }
      11. SelectCourse:   
      {
         "courseName":
      }
      12. DropCourse:   
      {
         null
      }  
   3. Meta data: version of the API, version of the application, and so on.
2. Responses are also json objects in HTTP posts, which contains the following parts and a status code of 200(when backend works well):
   1. Register:
      {
         "ErrorMessage": "string, empty when no error",
      }
   2. Remove:
      {
         "ErrorMessage": "string, empty when no error",
      }
   3. LogIn:   
      {
         "authToken": "string, unique for each user",
         "ErrorMessage": "string, empty when no error",
      }
   4. LogOut:   
      {
         "ErrorMessage": "string, empty when no error",
      }
   5. ModifyPassword:   
      {
         "ErrorMessage": "string, empty when no error",
      }
   6. GetAllUsersInfo:   
      {
         "users": [
            {
               "name": "string",
               "password": "string",
               "identityInfo": {
                  "class": {"grade": int, "class": int},
                  "previledge": int
               }
            },
            ...
         ],
      }
   7. GetPartUsersInfo:   
      {
         "users": [
            {
               "name": "string",
               "password": "string",
               "identityInfo": {
                  "class": {"grade": int, "class": int},
                  "previledge": int
               }
            },
            ...
         ],
      }
   8. AddCourse:   
      {
         "ErrorMessage": "string, empty when no error",
      }
   9. ModifyCourse:   
      {
         "ErrorMessage": "string, empty when no error",
      }
   10. LaunchCourse:   
      {
         "ErrorMessage": "string, empty when no error",
      }
   11. SelectCourse:   
      {
         "ErrorMessage": "string, empty when no error",
      }
   12. DropCourse:   
      {
         "ErrorMessage": "string, empty when no error",
      } 

### More Specifc Design and Implementation
Please view .md files in docs/. 