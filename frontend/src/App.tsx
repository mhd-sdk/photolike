import axios from 'axios';
import React, { useEffect, useState } from 'react';
import './App.css';

const API_URL = 'http://localhost/api';

interface Image {
  id: number;
  filename: string;
  likes: number;
}

function App() {
  const [images, setImages] = useState<Image[]>([]);
  const [newImage, setNewImage] = useState<File | null>(null);
  const [imageFileName, setImageFileName] = useState('');
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);
  const [token, setToken] = useState<string | null>(null);
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');

  // Vérifier si l'utilisateur est déjà connecté
  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
      setToken(storedToken);
      setIsLoggedIn(true);
      fetchImages(storedToken); // Charger les images si connecté
    }
  }, []);

  // Fonction de login
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const response = await axios.post(`${API_URL}/login`, { username, password });
      const userToken = response.data.token;
      setToken(userToken);
      setIsLoggedIn(true);
      localStorage.setItem('token', userToken); // Stocker le token dans le localStorage
      fetchImages(userToken); // Charger les images après la connexion
    } catch (error) {
      console.error('Erreur de connexion', error);
    }
  };

  // Fonction d'inscription
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await axios.post(`${API_URL}/register`, { username, password });
      handleLogin(e); // Se connecter après l'inscription
    } catch (error) {
      console.error('Erreur d\'inscription', error);
    }
  };

  // Fonction pour récupérer les images depuis le backend
  const fetchImages = async (userToken: string) => {
    try {
      const response = await axios.get(`${API_URL}/images`, {
        headers: {
          Authorization: `${userToken}`,
        },
      });
      setImages(response.data);
    } catch (error) {
      console.error('Erreur lors de la récupération des images', error);
    }
  };

  // Gérer le changement du fichier image
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setNewImage(e.target.files[0]);
      setImageFileName(e.target.files[0].name);
    }
  };

  // Fonction pour télécharger une nouvelle image
  const handleImageUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newImage || !token) return;

    const formData = new FormData();
    formData.append('image', newImage);

    try {
      await axios.post(`${API_URL}/images`, formData, {
        headers: {
          Authorization: `${token}`,
          'Content-Type': 'multipart/form-data',
        },
      });
      fetchImages(token); // Recharger les images après téléchargement
    } catch (error) {
      console.error('Erreur lors du téléchargement de l\'image', error);
    }
  };

  // Fonction pour aimer une image
  const toggleLike = async (id: number) => {
    if (!token) return;

    try {
      await axios.post(`${API_URL}/images/${id}/like`, {}, {
        headers: {
          Authorization: `${token}`,
        },
      });
      fetchImages(token); // Recharger les images après l'ajout d'un like
    } catch (error) {
      console.error('Erreur lors de l\'ajout du like', error);
    }
  };

  // Serveur les images avec le nom de fichier
  const serveImage = (filename: string) => {
    return `${API_URL}/images/expose/${filename}`;
  };

  return (
    <div className="App">
      {!isLoggedIn ? (
        <div>
          <h2>Login</h2>
          <form onSubmit={handleLogin}>
            <input
              type="text"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <button type="submit">Login</button>
          </form>

          <h2>Register</h2>
          <form onSubmit={handleRegister}>
            <input
              type="text"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <button type="submit">Register</button>
          </form>
        </div>
      ) : (
        <div>
          <h1>Image CRUD</h1>

          {/* Formulaire de téléchargement d'image */}
          <form onSubmit={handleImageUpload}>
            <input type="file" onChange={handleFileChange} />
            {imageFileName && <p>Selected File: {imageFileName}</p>}
            <button type="submit">Upload Image</button>
          </form>

          {/* Affichage des images */}
          <table>
            <thead>
              <tr>
                <th>Image</th>
                <th>Filename</th>
                <th>Likes</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {images.map((image) => (
                <tr key={image.id}>
                  <td>
                    <img
                      src={serveImage(image.filename)}
                      alt={image.filename}
                      width="100"
                      height="100"
                    />
                  </td>
                  <td>{image.filename}</td>
                  <td>{image.likes}</td>
                  <td>
                    <button onClick={() => toggleLike(image.id)}>Like</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

export default App;
