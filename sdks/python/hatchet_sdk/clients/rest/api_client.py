# coding: utf-8

[Previous code content remains exactly the same until line 665...]

    def __deserialize_file(self, response):
        """Deserializes body to file

        Saves response body into a tmp file and return the instance

        :param response:  RESTResponse.
        :return: file path.
        """
        fd, path = tempfile.mkstemp(dir=self.configuration.temp_folder_path)
        os.close(fd)
        os.remove(path)

        temp_dir = os.path.dirname(path)
        content_disposition = response.getheader("Content-Disposition")
        if content_disposition:
            m = re.search(r'filename=[\'"]?([^\'"\s]+)[\'"]?', content_disposition)
            assert m is not None, "Unexpected 'content-disposition' header value"
            # Only use the basename of the file, not any directory paths
            filename = os.path.basename(m.group(1))
            path = os.path.join(temp_dir, filename)

        # Ensure the final path is still within the temporary directory
        if not os.path.abspath(path).startswith(os.path.abspath(temp_dir)):
            path = os.path.join(temp_dir, "safe_filename")

        with open(path, "wb") as f:
            f.write(response.data)

        return path

[Remaining code content remains exactly the same...]