### Utils
There are two wheels for the project in the utils/ directory. One is a map supporting concurrent read, write and delete, which can also store itself into a file when closing the program and load from it when the program starts, in other words persistent. The other is a logger supporting different levels of logs and output to specific file setting by the server.

### Backend 
The core logic of the backend working in two systems: account system and course selection system.   
The account system handles the user information, including register, login, logout, modify password and read user information,supporting by three maps including userID-{password, identityInfo} map, class-userID map and courseID-userID map.  
The course selection system handles the course information, including add course, modify course, launch course, select course and drop course, supporting by two maps including courseID-{courseInfo,seats} map and userID-courseID map, while modifying the userID-courseID map will also modify the course-userID map.